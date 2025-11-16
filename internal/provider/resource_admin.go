package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &adminResource{}

func NewAdminResource() resource.Resource { return &adminResource{} }

type adminResource struct{ client *client.Client }

type adminResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Provisioner types.String `tfsdk:"provisioner"`
}

func (r *adminResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_admin"
}

func (r *adminResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Required: true},
			"provisioner": schema.StringAttribute{Required: true},
		},
	}
}

func (r *adminResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		r.client = c
	}
}

func (r *adminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data adminResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	a := client.Admin{Name: data.Name.ValueString(), Provisioner: data.Provisioner.ValueString()}
	if err := r.client.CreateAdmin(ctx, a); err != nil {
		resp.Diagnostics.AddError("create failed", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *adminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data adminResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	a, err := r.client.GetAdmin(ctx, data.Name.ValueString(), data.Provisioner.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read failed", err.Error())
		return
	}
	if a == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Provisioner = types.StringValue(a.Provisioner)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *adminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan adminResourceModel
	var state adminResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	plannedAdmin := client.Admin{Name: plan.Name.ValueString(), Provisioner: plan.Provisioner.ValueString()}
	if plannedAdmin.Name == state.Name.ValueString() && plannedAdmin.Provisioner == state.Provisioner.ValueString() {
		diags = resp.State.Set(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		return
	}
	if err := r.client.UpdateAdmin(ctx, state.Name.ValueString(), state.Provisioner.ValueString(), plannedAdmin); err != nil {
		resp.Diagnostics.AddError("update failed", err.Error())
		return
	}
	updated, err := r.client.GetAdmin(ctx, plannedAdmin.Name, plannedAdmin.Provisioner)
	if err != nil {
		resp.Diagnostics.AddError("refresh failed", err.Error())
		return
	}
	if updated == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	plan.Name = types.StringValue(updated.Name)
	plan.Provisioner = types.StringValue(updated.Provisioner)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *adminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data adminResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	if err := r.client.DeleteAdmin(ctx, data.Name.ValueString(), data.Provisioner.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete failed", err.Error())
		return
	}
}
