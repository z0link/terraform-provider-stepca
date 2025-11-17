package provider

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
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

func TestAdminResourceUpdateProvisionerChange(t *testing.T) {
	t.Parallel()
	fake := &fakeAdminClient{
		getAdminResp: &client.Admin{Name: "alice", Provisioner: "operators"},
	}
	r := &adminResource{client: fake}
	plan := adminResourceModel{Name: types.StringValue("alice"), ProvisionerName: types.StringValue("operators")}
	state := adminResourceModel{Name: types.StringValue("alice"), ProvisionerName: types.StringValue("admin")}
	updated, diags := r.updateAdmin(context.Background(), &state, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !fake.replaceCalled {
		t.Fatalf("expected replace to be called")
	}
	if updated == nil || updated.ProvisionerName.ValueString() != "operators" {
		t.Fatalf("unexpected updated state: %#v", updated)
	}
}

func TestAdminResourceUpdateNameImmutable(t *testing.T) {
	t.Parallel()
	fake := &fakeAdminClient{}
	r := &adminResource{client: fake}
	plan := adminResourceModel{Name: types.StringValue("bob"), ProvisionerName: types.StringValue("admin")}
	state := adminResourceModel{Name: types.StringValue("alice"), ProvisionerName: types.StringValue("admin")}
	updated, diags := r.updateAdmin(context.Background(), &state, &plan)
	if updated != nil {
		t.Fatalf("expected no updated state")
	}
	if !diags.HasError() {
		t.Fatalf("expected diagnostics")
	}
	if got := diags[0].Summary(); got != "name is immutable" {
		t.Fatalf("unexpected diagnostic summary: %s", got)
	}
	if fake.replaceCalled {
		t.Fatalf("replace should not be called when name changes")
	}
}

type fakeAdminClient struct {
	replaceCalled bool
	getAdminResp  *client.Admin
}

func (f *fakeAdminClient) CreateAdmin(ctx context.Context, a client.Admin) error { return nil }

func (f *fakeAdminClient) ReplaceAdmin(ctx context.Context, currentName, currentProvisioner string, a client.Admin) error {
	f.replaceCalled = true
	return nil
}

func (f *fakeAdminClient) DeleteAdmin(ctx context.Context, name, provisioner string) error {
	return nil
}

func (f *fakeAdminClient) GetAdmin(ctx context.Context, name, provisioner string) (*client.Admin, error) {
	return f.getAdminResp, nil
}
