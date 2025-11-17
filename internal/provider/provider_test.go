package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateAdminCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		model   stepcaProviderModel
		wantErr bool
		summary string
	}{
		{
			name: "token only",
			model: stepcaProviderModel{
				AdminToken: types.StringValue("token"),
			},
		},
		{
			name: "admin key pair",
			model: stepcaProviderModel{
				AdminName: types.StringValue("admin@example.com"),
				AdminKey:  types.StringValue("/path/to/key"),
			},
		},
		{
			name: "token and key pair",
			model: stepcaProviderModel{
				AdminToken: types.StringValue("token"),
				AdminName:  types.StringValue("admin@example.com"),
				AdminKey:   types.StringValue("/path/to/key"),
			},
		},
		{
			name:    "missing all",
			model:   stepcaProviderModel{},
			wantErr: true,
			summary: "missing admin credentials",
		},
		{
			name: "name without key",
			model: stepcaProviderModel{
				AdminName: types.StringValue("admin@example.com"),
			},
			wantErr: true,
			summary: "incomplete admin key configuration",
		},
		{
			name: "key without name",
			model: stepcaProviderModel{
				AdminKey: types.StringValue("/path/to/key"),
			},
			wantErr: true,
			summary: "incomplete admin key configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateAdminCredentials(&tt.model)
			if diags.HasError() != tt.wantErr {
				t.Fatalf("unexpected error state: %v", diags)
			}
			if tt.wantErr {
				if len(diags) == 0 {
					t.Fatalf("expected diagnostic")
				}
				if got := diags[0].Summary(); got != tt.summary {
					t.Fatalf("unexpected summary: %s", got)
				}
			}
		})
	}
}
