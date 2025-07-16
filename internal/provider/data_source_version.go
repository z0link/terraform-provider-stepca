package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ datasource.DataSource = &versionDataSource{}

func NewVersionDataSource() datasource.DataSource {
	return &versionDataSource{}
}

type versionDataSource struct {
	client *client.Client
}

type versionDataSourceModel struct {
	Version types.String `tfsdk:"version"`
}

func (d *versionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "stepca_version"
}

func (d *versionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *versionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		d.client = c
	}
}

func (d *versionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	version, err := d.client.Version(ctx)
	if err != nil {
		resp.Diagnostics.AddError("version fetch failed", err.Error())
		return
	}

	data := versionDataSourceModel{Version: types.StringValue(version)}
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
