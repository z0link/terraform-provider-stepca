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
	CAURL            types.String `tfsdk:"ca_url"`
	AdminName        types.String `tfsdk:"admin_name"`
	AdminKey         types.String `tfsdk:"admin_key"`
	AdminProvisioner types.String `tfsdk:"admin_provisioner"`
	Token            types.String `tfsdk:"token"`
	AdminToken       types.String `tfsdk:"admin_token"`
}

func (p *stepcaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stepca"
	resp.Version = version.Version
}

func (p *stepcaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ca_url":            schema.StringAttribute{Required: true},
			"admin_name":        schema.StringAttribute{Required: true},
			"admin_key":         schema.StringAttribute{Required: true},
			"admin_provisioner": schema.StringAttribute{Optional: true},
			"token":             schema.StringAttribute{Required: true, Sensitive: true},
			// Token for admin API calls. Generate with the admin key if
			// using a JWK admin provisioner.
			"admin_token": schema.StringAttribute{Optional: true, Sensitive: true},
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
	c = c.WithAdminName(data.AdminName.ValueString())
	c = c.WithAdminKey(data.AdminKey.ValueString())
	if !data.AdminProvisioner.IsNull() {
		c = c.WithAdminProvisioner(data.AdminProvisioner.ValueString())
	}
	if !data.AdminToken.IsNull() {
		c = c.WithAdminToken(data.AdminToken.ValueString())
	}
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *stepcaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCertificateResource,
		NewProvisionerResource,
		NewAdminResource,
	}
}

func (p *stepcaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
        return []func() datasource.DataSource{
                NewVersionDataSource,
                NewCACertificateDataSource,
                NewProvisionersDataSource,
        }
}
