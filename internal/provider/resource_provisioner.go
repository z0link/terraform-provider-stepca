package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &provisionerResource{}

func NewProvisionerResource() resource.Resource {
	return &provisionerResource{}
}

type provisionerResource struct {
	client *client.Client
}

type provisionerResourceModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Admin types.Bool   `tfsdk:"admin"`
}

func (r *provisionerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_provisioner"
}

func (r *provisionerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":  schema.StringAttribute{Required: true},
			"type":  schema.StringAttribute{Required: true},
			"admin": schema.BoolAttribute{Optional: true},
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
	p := client.Provisioner{Name: data.Name.ValueString(), Type: data.Type.ValueString(), Admin: !data.Admin.IsNull() && data.Admin.ValueBool()}
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
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *provisionerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	p := client.Provisioner{Name: data.Name.ValueString(), Type: data.Type.ValueString(), Admin: !data.Admin.IsNull() && data.Admin.ValueBool()}
	if err := r.client.CreateProvisioner(ctx, p); err != nil {
		resp.Diagnostics.AddError("update failed", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
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
