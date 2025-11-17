package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

type fakeProvisionerClient struct {
	replaceCalled bool
	replaceInput  client.Provisioner
	getResp       *client.Provisioner
}

func (f *fakeProvisionerClient) CreateProvisioner(ctx context.Context, p client.Provisioner) error {
	return nil
}

func (f *fakeProvisionerClient) ReplaceProvisioner(ctx context.Context, name string, p client.Provisioner) error {
	f.replaceCalled = true
	f.replaceInput = p
	return nil
}

func (f *fakeProvisionerClient) DeleteProvisioner(ctx context.Context, name string) error { return nil }

func (f *fakeProvisionerClient) GetProvisioner(ctx context.Context, name string) (*client.Provisioner, error) {
	return f.getResp, nil
}

func TestProvisionerResourceUpdateAdminToggle(t *testing.T) {
	t.Parallel()
	fake := &fakeProvisionerClient{getResp: &client.Provisioner{Name: "api", Type: "JWK", Admin: true}}
	r := &provisionerResource{client: fake}
	plan := provisionerResourceModel{Name: types.StringValue("api"), Type: types.StringValue("JWK"), Admin: types.BoolValue(true)}
	state := provisionerResourceModel{Name: types.StringValue("api"), Type: types.StringValue("JWK"), Admin: types.BoolValue(false)}
	updated, diags := r.updateProvisioner(context.Background(), &state, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !fake.replaceCalled {
		t.Fatalf("expected replace call")
	}
	if updated == nil || !updated.Admin.ValueBool() {
		t.Fatalf("unexpected updated state: %#v", updated)
	}
	if fake.replaceInput.Admin != true {
		t.Fatalf("unexpected replace payload: %#v", fake.replaceInput)
	}
}

func TestProvisionerResourceUpdateTypeImmutable(t *testing.T) {
	t.Parallel()
	fake := &fakeProvisionerClient{}
	r := &provisionerResource{client: fake}
	plan := provisionerResourceModel{Name: types.StringValue("api"), Type: types.StringValue("JWK"), Admin: types.BoolValue(true)}
	state := provisionerResourceModel{Name: types.StringValue("api"), Type: types.StringValue("OIDC"), Admin: types.BoolValue(false)}
	updated, diags := r.updateProvisioner(context.Background(), &state, &plan)
	if updated != nil {
		t.Fatalf("expected nil updated state")
	}
	if !diags.HasError() {
		t.Fatalf("expected diagnostics")
	}
	if got := diags[0].Summary(); got != "type is immutable" {
		t.Fatalf("unexpected diagnostic summary: %s", got)
	}
	if fake.replaceCalled {
		t.Fatalf("replace should not be called when type changes")
	}
}
