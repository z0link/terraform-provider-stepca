package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ datasource.DataSource = &caCertificateDataSource{}

func NewCACertificateDataSource() datasource.DataSource {
	return &caCertificateDataSource{}
}

type caCertificateDataSource struct {
	client *client.Client
}

type caCertificateDataSourceModel struct {
	Certificate types.String `tfsdk:"certificate"`
}

func (d *caCertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "stepca_ca_certificate"
}

func (d *caCertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"certificate": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *caCertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		d.client = c
	}
}

func (d *caCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}
	pem, err := d.client.RootCertificate(ctx)
	if err != nil {
		resp.Diagnostics.AddError("fetch failed", err.Error())
		return
	}
	data := caCertificateDataSourceModel{Certificate: types.StringValue(string(pem))}
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
