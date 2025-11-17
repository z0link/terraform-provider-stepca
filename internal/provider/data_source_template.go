package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ datasource.DataSource = &templateDataSource{}

func NewTemplateDataSource() datasource.DataSource {
	return &templateDataSource{}
}

type templateDataSource struct {
	client templateGetter
}

type templateDataSourceModel struct {
	Name     types.String `tfsdk:"name"`
	Body     types.String `tfsdk:"body"`
	Metadata types.Map    `tfsdk:"metadata"`
}

func (d *templateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "stepca_template"
}

func (d *templateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":     schema.StringAttribute{Required: true},
			"body":     schema.StringAttribute{Computed: true},
			"metadata": schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func (d *templateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		d.client = c
	}
}

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	var data templateDataSourceModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, metadata, helperDiags := GetTemplate(ctx, d.client, data.Name.ValueString())
	resp.Diagnostics.Append(helperDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Body = types.StringValue(body)
	data.Metadata = metadata

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
