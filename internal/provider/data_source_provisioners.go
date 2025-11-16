package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

var _ datasource.DataSource = &provisionersDataSource{}

func NewProvisionersDataSource() datasource.DataSource {
	return &provisionersDataSource{}
}

type provisionersDataSource struct {
	client *client.Client
}

type provisionersDataSourceModel struct {
	Provisioners []provisionerItemModel `tfsdk:"provisioners"`
}

type provisionerItemModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Admin types.Bool   `tfsdk:"admin"`
}

func (d *provisionersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "stepca_provisioners"
}

func (d *provisionersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"provisioners": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Computed: true},
						"type":  schema.StringAttribute{Computed: true},
						"admin": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *provisionersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	if c, ok := req.ProviderData.(*client.Client); ok {
		d.client = c
	}
}

func (d *provisionersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client")
		return
	}

	items, err := d.client.ListProvisioners(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to list provisioners", err.Error())
		return
	}

	data := provisionersDataSourceModel{Provisioners: make([]provisionerItemModel, 0, len(items))}
	for _, item := range items {
		data.Provisioners = append(data.Provisioners, provisionerItemModel{
			Name:  types.StringValue(item.Name),
			Type:  types.StringValue(item.Type),
			Admin: types.BoolValue(item.Admin),
		})
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
