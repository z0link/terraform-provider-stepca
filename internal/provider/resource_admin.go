package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &adminResource{}

func NewAdminResource() resource.Resource { return &adminResource{} }

type adminClient interface {
	CreateAdmin(ctx context.Context, a client.Admin) error
	ReplaceAdmin(ctx context.Context, currentName, currentProvisioner string, a client.Admin) error
	DeleteAdmin(ctx context.Context, name, provisioner string) error
	GetAdmin(ctx context.Context, name, provisioner string) (*client.Admin, error)
}

type adminResource struct{ client adminClient }

type adminResourceModel struct {
	Name            types.String `tfsdk:"name"`
	ProvisionerName types.String `tfsdk:"provisioner_name"`
}

func (r *adminResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_admin"
}

func (r *adminResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":             schema.StringAttribute{Required: true},
			"provisioner_name": schema.StringAttribute{Required: true},
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
	a := client.Admin{Name: data.Name.ValueString(), Provisioner: data.ProvisionerName.ValueString()}
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
	a, err := r.client.GetAdmin(ctx, data.Name.ValueString(), data.ProvisionerName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read failed", err.Error())
		return
	}
	if a == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(a.Name)
	data.ProvisionerName = types.StringValue(a.Provisioner)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *adminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan adminResourceModel
	var state adminResourceModel
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
	updated, updateDiags := r.updateAdmin(ctx, &state, &plan)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, updated)
	resp.Diagnostics.Append(diags...)
}

func (r *adminResource) updateAdmin(ctx context.Context, state, plan *adminResourceModel) (*adminResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.client == nil {
		diags.AddError("provider not configured", "missing client")
		return nil, diags
	}
	if plan.Name.ValueString() != state.Name.ValueString() {
		diags.AddError("name is immutable", "changing the admin name requires recreating the resource")
		return nil, diags
	}
	if plan.ProvisionerName.ValueString() != state.ProvisionerName.ValueString() {
		payload := client.Admin{Name: plan.Name.ValueString(), Provisioner: plan.ProvisionerName.ValueString()}
		if err := r.client.ReplaceAdmin(ctx, state.Name.ValueString(), state.ProvisionerName.ValueString(), payload); err != nil {
			diags.AddError("update failed", err.Error())
			return nil, diags
		}
	}
	updated, err := r.client.GetAdmin(ctx, plan.Name.ValueString(), plan.ProvisionerName.ValueString())
	if err != nil {
		diags.AddError("read failed", err.Error())
		return nil, diags
	}
	if updated == nil {
		diags.AddError("read failed", "admin missing after update")
		return nil, diags
	}
	result := &adminResourceModel{
		Name:            types.StringValue(updated.Name),
		ProvisionerName: types.StringValue(updated.Provisioner),
	}
	return result, diags
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
	if err := r.client.DeleteAdmin(ctx, data.Name.ValueString(), data.ProvisionerName.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete failed", err.Error())
		return
	}
}
