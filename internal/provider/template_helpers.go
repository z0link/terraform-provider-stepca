package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetTemplate fetches a template body and metadata map from the admin API.
func GetTemplate(ctx context.Context, getter templateGetter, name string) (string, types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	tmpl, err := getter.GetTemplate(ctx, name)
	if err != nil {
		diags = append(diags, diag.NewErrorDiagnostic("get template failed", err.Error()))
		return "", types.MapNull(types.StringType), diags
	}
	if tmpl == nil {
		diags = append(diags, diag.NewErrorDiagnostic("template not found", fmt.Sprintf("template %q was not found", name)))
		return "", types.MapNull(types.StringType), diags
	}

	metadata := types.MapNull(types.StringType)
	if len(tmpl.Metadata) > 0 {
		val, mapDiags := types.MapValueFrom(ctx, types.StringType, tmpl.Metadata)
		diags = append(diags, mapDiags...)
		if diags.HasError() {
			return tmpl.Body, metadata, diags
		}
		metadata = val
	}

	return tmpl.Body, metadata, diags
}
