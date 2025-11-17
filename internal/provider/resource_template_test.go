package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTemplateResourceSchema(t *testing.T) {
	t.Parallel()

	resource := NewTemplateResource()
	var resp pfresource.SchemaResponse
	resource.Schema(context.Background(), pfresource.SchemaRequest{}, &resp)

	expected := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":     schema.StringAttribute{Required: true},
			"body":     schema.StringAttribute{Required: true},
			"metadata": schema.MapAttribute{Optional: true, ElementType: types.StringType},
		},
	}

	if diff := cmp.Diff(expected, resp.Schema); diff != "" {
		t.Fatalf("unexpected schema: (-want +got)\n%s", diff)
	}
}
