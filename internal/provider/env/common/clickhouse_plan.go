package env

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// A plan can still hold unknown nested objects (disk = var.x), which a struct target refuses.
func ObjectsFromMap[M any](ctx context.Context, m types.Map) (map[string]M, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}

	lenient := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	models := make(map[string]M, len(m.Elements()))
	for k, e := range m.Elements() {
		obj, ok := e.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			continue
		}
		var model M
		diags.Append(obj.As(ctx, &model, lenient)...)
		models[k] = model
	}

	return models, diags
}

// The API refuses a new user without a password, and silently ignores a password type change without one.
func ValidateClickHouseUserPasswords(root path.Path, prior, planned map[string]ClickHouseClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for clusterName, c := range planned {
		for userName, u := range c.Users {
			if u.Password == nil || !u.Password.EncValue.IsNull() {
				continue
			}
			encValue := root.AtMapKey(clusterName).AtName("users").AtMapKey(userName).AtName("password").AtName("enc_value")
			priorUser, exists := prior[clusterName].Users[userName]
			switch {
			case !exists:
				diags.AddAttributeError(encValue, "Missing ClickHouse User Password",
					fmt.Sprintf("User %q in cluster %q does not exist yet, and the API requires a password to create it. Set enc_value; it can be omitted later to keep the current password.", userName, clusterName))
			case priorUser.Password != nil && !priorUser.Password.Type.IsNull() && !u.Password.Type.IsUnknown() && !u.Password.Type.Equal(priorUser.Password.Type):
				diags.AddAttributeError(encValue, "Missing ClickHouse User Password",
					fmt.Sprintf("User %q in cluster %q changes its password type from %s to %s, which needs a new password in that form. Set enc_value.", userName, clusterName, priorUser.Password.Type.ValueString(), u.Password.Type.ValueString()))
			}
		}
	}

	return diags
}
