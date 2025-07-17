package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
	"github.com/z0link/terraform-provider-stepca/internal/version"
)

// Ensure implementation satisfies interfaces
var _ provider.Provider = &stepcaProvider{}

// New creates a new provider
func New() provider.Provider {
	return &stepcaProvider{}
}

type stepcaProvider struct{}

type stepcaProviderModel struct {
	CAURL types.String `tfsdk:"ca_url"`
	Token types.String `tfsdk:"token"`
}

func (p *stepcaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stepca"
	resp.Version = version.Version
}

func (p *stepcaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ca_url": schema.StringAttribute{Required: true},
			"token":  schema.StringAttribute{Required: true, Sensitive: true},
		},
	}
}

func (p *stepcaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data stepcaProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(data.CAURL.ValueString(), data.Token.ValueString())
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *stepcaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCertificateResource,
	}
}

func (p *stepcaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVersionDataSource,
	}
}
