package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

func TestTemplateDataSourceSchema(t *testing.T) {
	t.Parallel()

	ds := NewTemplateDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	expected := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":     schema.StringAttribute{Required: true},
			"body":     schema.StringAttribute{Computed: true},
			"metadata": schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}

	if diff := cmp.Diff(expected, resp.Schema); diff != "" {
		t.Fatalf("unexpected schema: (-want +got)\n%s", diff)
	}
}

func TestGetTemplateMetadata(t *testing.T) {
	t.Parallel()

	fake := templateGetterFunc(func(ctx context.Context, name string) (*client.Template, error) {
		return &client.Template{
			Name: name,
			Body: "body",
			Metadata: map[string]string{
				"team": "platform",
			},
		}, nil
	})

	body, metadata, diags := GetTemplate(context.Background(), fake, "example")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if body != "body" {
		t.Fatalf("unexpected body %q", body)
	}

	var got map[string]string
	if err := metadata.ElementsAs(context.Background(), &got, false); err != nil {
		t.Fatalf("failed to decode metadata: %v", err)
	}
	want := map[string]string{"team": "platform"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected metadata: (-want +got)\n%s", diff)
	}
}

type templateGetterFunc func(context.Context, string) (*client.Template, error)

func (f templateGetterFunc) GetTemplate(ctx context.Context, name string) (*client.Template, error) {
	return f(ctx, name)
}
