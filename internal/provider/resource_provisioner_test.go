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

func TestProvisionerResourceSchema(t *testing.T) {
	t.Parallel()
	resource := NewProvisionerResource()
	var resp pfresource.SchemaResponse
	resource.Schema(context.Background(), pfresource.SchemaRequest{}, &resp)
	expected := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":                 schema.StringAttribute{Required: true},
			"type":                 schema.StringAttribute{Required: true},
			"admin":                schema.BoolAttribute{Optional: true},
			"x509_template":        schema.StringAttribute{Optional: true},
			"ssh_template":         schema.StringAttribute{Optional: true},
			"attestation_template": schema.StringAttribute{Optional: true},
		},
	}
	if diff := cmp.Diff(expected, resp.Schema); diff != "" {
		t.Fatalf("unexpected schema: (-want +got)\n%s", diff)
	}
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

func TestProvisionerResourceUpdateTemplateChange(t *testing.T) {
	t.Parallel()
	fake := &fakeProvisionerClient{getResp: &client.Provisioner{Name: "api", Type: "JWK", X509Template: "leaf"}}
	r := &provisionerResource{client: fake}
	plan := provisionerResourceModel{
		Name:         types.StringValue("api"),
		Type:         types.StringValue("JWK"),
		X509Template: types.StringValue("leaf"),
		SSHTemplate:  types.StringValue("ssh-user"),
	}
	state := provisionerResourceModel{
		Name:         types.StringValue("api"),
		Type:         types.StringValue("JWK"),
		X509Template: types.StringNull(),
		SSHTemplate:  types.StringNull(),
	}
	updated, diags := r.updateProvisioner(context.Background(), &state, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !fake.replaceCalled {
		t.Fatalf("expected replace call")
	}
	if fake.replaceInput.X509Template != "leaf" || fake.replaceInput.SSHTemplate != "ssh-user" {
		t.Fatalf("unexpected template payload: %#v", fake.replaceInput)
	}
	if updated == nil || updated.X509Template.IsNull() {
		t.Fatalf("expected updated model to include template: %#v", updated)
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
