package provider

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &certificateResource{}

func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

type certificateClient interface {
	Sign(ctx context.Context, csr string) ([]byte, error)
	Certificate(ctx context.Context, serial string) ([]byte, bool, error)
}

type certificateResource struct {
	client certificateClient
}

type certificateResourceModel struct {
	CSR         types.String `tfsdk:"csr"`
	Cert        types.String `tfsdk:"certificate"`
	ForceRotate types.Bool   `tfsdk:"force_rotate"`
}

func (r *certificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_certificate"
}

func (r *certificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"csr":         schema.StringAttribute{Required: true},
			"certificate": schema.StringAttribute{Computed: true},
			"force_rotate": schema.BoolAttribute{
				Optional:    true,
				Description: "Toggle this value to force Terraform to request a new certificate without changing the CSR.",
			},
		},
	}
}

func (r *certificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		r.client = c
	}
}

func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data certificateResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	certPEM, err := r.client.Sign(ctx, data.CSR.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("sign failed", err.Error())
		return
	}
	data.Cert = types.StringValue(string(certPEM))
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data certificateResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	keep, readDiags := r.shouldKeepCertificate(ctx, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !keep {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state certificateResourceModel
	var plan certificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, updateDiags := r.applyCertificateUpdate(ctx, plan, state)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &updated)
	resp.Diagnostics.Append(diags...)
}

func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *certificateResource) shouldKeepCertificate(ctx context.Context, data *certificateResourceModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if data.Cert.IsNull() || data.Cert.IsUnknown() || data.Cert.ValueString() == "" {
		diags = append(diags, diag.NewWarningDiagnostic(
			"certificate missing",
			"The certificate value is empty in state. Removing it so Terraform can request a new certificate.",
		))
		return false, diags
	}

	if r.client == nil {
		return true, diags
	}

	cert, err := parseCertificate(data.Cert.ValueString())
	if err != nil {
		diags = append(diags, diag.NewWarningDiagnostic(
			"certificate parse failed",
			fmt.Sprintf("The stored certificate could not be parsed (%v). Removing it so Terraform can recreate the resource.", err),
		))
		return false, diags
	}

	serial := strings.ToLower(cert.SerialNumber.Text(16))
	remotePEM, found, err := r.client.Certificate(ctx, serial)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("certificate lookup failed", err.Error()))
		return true, diags
	}

	if !found {
		diags = append(diags, diag.NewWarningDiagnostic(
			"certificate revoked",
			"The certificate could not be located via the CA API. Removing it from state so Terraform can issue a new one.",
		))
		return false, diags
	}

	remoteCert, err := parseCertificate(string(remotePEM))
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("certificate parse failed", err.Error()))
		return true, diags
	}

	if !bytes.Equal(remoteCert.Raw, cert.Raw) {
		diags = append(diags, diag.NewWarningDiagnostic(
			"certificate drift detected",
			"The CA returned a different certificate for the stored serial number. Removing it from state so Terraform can request a new certificate.",
		))
		return false, diags
	}

	return true, diags
}

func (r *certificateResource) applyCertificateUpdate(ctx context.Context, plan, state certificateResourceModel) (certificateResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !needsCertificateRotation(plan, state) {
		plan.Cert = state.Cert
		return plan, diags
	}

	if r.client == nil {
		diags = append(diags, diag.NewErrorDiagnostic("provider not configured", "missing client"))
		return plan, diags
	}

	certPEM, err := r.client.Sign(ctx, plan.CSR.ValueString())
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("sign failed", err.Error()))
		return plan, diags
	}

	plan.Cert = types.StringValue(string(certPEM))
	return plan, diags
}

func needsCertificateRotation(plan, state certificateResourceModel) bool {
	if plan.CSR.ValueString() != state.CSR.ValueString() {
		return true
	}
	return boolValue(plan.ForceRotate) != boolValue(state.ForceRotate)
}

func boolValue(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return v.ValueBool()
}

func parseCertificate(pemData string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}
