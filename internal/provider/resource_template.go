package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ resource.Resource = &templateResource{}

// templateGetter captures the helper interface for retrieving templates by name.
type templateGetter interface {
	GetTemplate(context.Context, string) (*client.Template, error)
}

// templateClient captures the subset of the client API this resource needs.
type templateClient interface {
	templateGetter
	CreateTemplate(context.Context, client.Template) error
	UpdateTemplate(context.Context, client.Template) error
	DeleteTemplate(context.Context, string) error
}

func NewTemplateResource() resource.Resource { return &templateResource{} }

type templateResource struct{ client templateClient }

type templateResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Body     types.String `tfsdk:"body"`
	Metadata types.Map    `tfsdk:"metadata"`
}

func (r *templateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "stepca_template"
}

func (r *templateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Required: true},
			"body": schema.StringAttribute{Required: true},
			"metadata": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *templateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		r.client = c
	}
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	var data templateResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, convDiags := modelToTemplate(ctx, data)
	resp.Diagnostics.Append(convDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CreateTemplate(ctx, tmpl); err != nil {
		resp.Diagnostics.AddError("create failed", err.Error())
		return
	}

	if diags := setMetadataState(ctx, &data, tmpl.Metadata); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	var data templateResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.GetTemplate(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read failed", err.Error())
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Body = types.StringValue(tmpl.Body)
	if diags := setMetadataState(ctx, &data, tmpl.Metadata); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	var data templateResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, convDiags := modelToTemplate(ctx, data)
	resp.Diagnostics.Append(convDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateTemplate(ctx, tmpl); err != nil {
		resp.Diagnostics.AddError("update failed", err.Error())
		return
	}

	if diags := setMetadataState(ctx, &data, tmpl.Metadata); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	var data templateResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTemplate(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete failed", err.Error())
	}
}

func modelToTemplate(ctx context.Context, data templateResourceModel) (client.Template, diag.Diagnostics) {
	metadata := map[string]string{}
	var diags diag.Diagnostics
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		diags = data.Metadata.ElementsAs(ctx, &metadata, false)
	}
	tmpl := client.Template{
		Name:     data.Name.ValueString(),
		Body:     data.Body.ValueString(),
		Metadata: metadata,
	}
	if len(metadata) == 0 {
		tmpl.Metadata = nil
	}
	return tmpl, diags
}

func setMetadataState(ctx context.Context, data *templateResourceModel, metadata map[string]string) diag.Diagnostics {
	if len(metadata) == 0 {
		data.Metadata = types.MapNull(types.StringType)
		return nil
	}
	val, diags := types.MapValueFrom(ctx, types.StringType, metadata)
	if diags.HasError() {
		return diags
	}
	data.Metadata = val
	return nil
}
