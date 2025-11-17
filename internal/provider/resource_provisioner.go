package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &provisionerResource{}

func NewProvisionerResource() resource.Resource {
	return &provisionerResource{}
}

type provisionerClient interface {
	CreateProvisioner(ctx context.Context, p client.Provisioner) error
	ReplaceProvisioner(ctx context.Context, name string, p client.Provisioner) error
	DeleteProvisioner(ctx context.Context, name string) error
	GetProvisioner(ctx context.Context, name string) (*client.Provisioner, error)
}

type provisionerResource struct {
	client provisionerClient
}

type provisionerResourceModel struct {
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Admin               types.Bool   `tfsdk:"admin"`
	X509Template        types.String `tfsdk:"x509_template"`
	SSHTemplate         types.String `tfsdk:"ssh_template"`
	AttestationTemplate types.String `tfsdk:"attestation_template"`
}

func (r *provisionerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_provisioner"
}

func (r *provisionerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":                 schema.StringAttribute{Required: true},
			"type":                 schema.StringAttribute{Required: true},
			"admin":                schema.BoolAttribute{Optional: true},
			"x509_template":        schema.StringAttribute{Optional: true},
			"ssh_template":         schema.StringAttribute{Optional: true},
			"attestation_template": schema.StringAttribute{Optional: true},
		},
	}
}

func (r *provisionerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		r.client = c
	}
}

func (r *provisionerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data provisionerResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	p := provisionerModelToClient(data)
	if err := r.client.CreateProvisioner(ctx, p); err != nil {
		resp.Diagnostics.AddError("create failed", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *provisionerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data provisionerResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	p, err := r.client.GetProvisioner(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read failed", err.Error())
		return
	}
	if p == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Type = types.StringValue(p.Type)
	data.Admin = types.BoolValue(p.Admin)
	data.X509Template = stringValueOrNull(p.X509Template)
	data.SSHTemplate = stringValueOrNull(p.SSHTemplate)
	data.AttestationTemplate = stringValueOrNull(p.AttestationTemplate)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *provisionerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan provisionerResourceModel
	var state provisionerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, updateDiags := r.updateProvisioner(ctx, &state, &plan)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, updated)
	resp.Diagnostics.Append(diags...)
}

func (r *provisionerResource) updateProvisioner(ctx context.Context, state, plan *provisionerResourceModel) (*provisionerResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.client == nil {
		diags.AddError("provider not configured", "missing client")
		return nil, diags
	}
	if plan.Name.ValueString() != state.Name.ValueString() {
		diags.AddError("name is immutable", "changing the name requires recreating the provisioner")
		return nil, diags
	}
	if plan.Type.ValueString() != state.Type.ValueString() {
		diags.AddError("type is immutable", "changing the type requires recreating the provisioner")
		return nil, diags
	}
	planAdmin := boolFromOptional(plan.Admin)
	stateAdmin := boolFromOptional(state.Admin)
	shouldReplace := planAdmin != stateAdmin ||
		!stringAttrEqual(plan.X509Template, state.X509Template) ||
		!stringAttrEqual(plan.SSHTemplate, state.SSHTemplate) ||
		!stringAttrEqual(plan.AttestationTemplate, state.AttestationTemplate)
	if shouldReplace {
		payload := provisionerModelToClient(*plan)
		if err := r.client.ReplaceProvisioner(ctx, state.Name.ValueString(), payload); err != nil {
			diags.AddError("update failed", err.Error())
			return nil, diags
		}
	}
	updated, err := r.client.GetProvisioner(ctx, plan.Name.ValueString())
	if err != nil {
		diags.AddError("read failed", err.Error())
		return nil, diags
	}
	if updated == nil {
		diags.AddError("read failed", "provisioner missing after update")
		return nil, diags
	}
	result := &provisionerResourceModel{
		Name:                types.StringValue(updated.Name),
		Type:                types.StringValue(updated.Type),
		Admin:               types.BoolValue(updated.Admin),
		X509Template:        stringValueOrNull(updated.X509Template),
		SSHTemplate:         stringValueOrNull(updated.SSHTemplate),
		AttestationTemplate: stringValueOrNull(updated.AttestationTemplate),
	}
	return result, diags
}

func provisionerModelToClient(data provisionerResourceModel) client.Provisioner {
	p := client.Provisioner{
		Name:  data.Name.ValueString(),
		Type:  data.Type.ValueString(),
		Admin: boolFromOptional(data.Admin),
	}
	if v, ok := optionalStringValue(data.X509Template); ok {
		p.X509Template = v
	}
	if v, ok := optionalStringValue(data.SSHTemplate); ok {
		p.SSHTemplate = v
	}
	if v, ok := optionalStringValue(data.AttestationTemplate); ok {
		p.AttestationTemplate = v
	}
	return p
}

func boolFromOptional(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return v.ValueBool()
}

func optionalStringValue(v types.String) (string, bool) {
	if v.IsNull() || v.IsUnknown() {
		return "", false
	}
	return v.ValueString(), true
}

func stringAttrEqual(a, b types.String) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsNull() != b.IsNull() {
		return false
	}
	if a.IsUnknown() || b.IsUnknown() {
		return false
	}
	return a.ValueString() == b.ValueString()
}

func stringValueOrNull(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func (r *provisionerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data provisionerResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	if err := r.client.DeleteProvisioner(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete failed", err.Error())
		return
	}
}
