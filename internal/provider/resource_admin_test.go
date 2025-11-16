package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestAdminResourceSchema(t *testing.T) {
	t.Parallel()

	resource := NewAdminResource()
	var resp pfresource.SchemaResponse
	resource.Schema(context.Background(), pfresource.SchemaRequest{}, &resp)

	expected := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":             schema.StringAttribute{Required: true},
			"provisioner_name": schema.StringAttribute{Required: true},
		},
	}

	if diff := cmp.Diff(expected, resp.Schema); diff != "" {
		t.Fatalf("unexpected schema: (-want +got)\n%s", diff)
	}
}
