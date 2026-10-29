package env

import (
	"context"
	"testing"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestObjectsFromMap(t *testing.T) {
	ctx := context.Background()
	elemType := clientsupport.GetClickHouseClustersAttribute(false, true, false).NestedObject.Type()

	t.Run("null and unknown maps read as nil", func(t *testing.T) {
		for _, m := range []types.Map{types.MapNull(elemType), types.MapUnknown(elemType)} {
			got, diags := ObjectsFromMap[ClickHouseClusterModel](ctx, m)
			if diags.HasError() || got != nil {
				t.Errorf("expected nil without diags, got %#v / %v", got, diags)
			}
		}
	})

	t.Run("an unknown nested object reads leniently", func(t *testing.T) {
		objType, ok := elemType.(types.ObjectType)
		if !ok {
			t.Fatalf("expected an object element type, got %T", elemType)
		}
		attrs := map[string]attr.Value{}
		for name, at := range objType.AttrTypes {
			attrs[name] = unknownValue(t, ctx, at)
		}
		attrs["instance_type"] = types.StringValue("m6i.large")
		obj := types.ObjectValueMust(objType.AttrTypes, attrs)
		m := types.MapValueMust(elemType, map[string]attr.Value{"ch": obj})

		got, diags := ObjectsFromMap[ClickHouseClusterModel](ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got["ch"].InstanceType.ValueString() != "m6i.large" {
			t.Errorf("expected the cluster keyed by name, got %#v", got)
		}
	})
}

func unknownValue(t *testing.T, ctx context.Context, at attr.Type) attr.Value {
	t.Helper()
	v, err := at.ValueFromTerraform(ctx, tftypes.NewValue(at.TerraformType(ctx), tftypes.UnknownValue))
	if err != nil {
		t.Fatalf("unknown %T: %v", at, err)
	}
	return v
}

func TestValidateClickHouseUserPasswords(t *testing.T) {
	user := func(encValue types.String) ClickHouseUserModel {
		return ClickHouseUserModel{
			Password: &ClickHouseUserPasswordModel{Type: types.StringValue("SHA256_HEX"), EncValue: encValue},
		}
	}
	cluster := func(users map[string]ClickHouseUserModel) ClickHouseClusterModel {
		return ClickHouseClusterModel{Users: users}
	}
	root := path.Root("clickhouse_clusters")

	t.Run("a new user needs a password", func(t *testing.T) {
		planned := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{"app": user(types.StringNull())})}
		diags := ValidateClickHouseUserPasswords(root, nil, planned)
		if !diags.HasError() {
			t.Fatal("expected an error for a new user without enc_value")
		}
		withPath, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
		if !ok {
			t.Fatalf("expected an attribute error, got %T", diags.Errors()[0])
		}
		want := root.AtMapKey("ch").AtName("users").AtMapKey("app").AtName("password").AtName("enc_value")
		if got := withPath.Path(); !got.Equal(want) {
			t.Errorf("error path = %s, want %s", got, want)
		}
	})

	t.Run("an existing user may omit it", func(t *testing.T) {
		prior := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{"app": user(types.StringValue("kept"))})}
		planned := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{"app": user(types.StringNull())})}
		if diags := ValidateClickHouseUserPasswords(root, prior, planned); diags.HasError() {
			t.Errorf("unexpected diags: %v", diags)
		}
	})

	t.Run("the same user name in a new cluster is a new user", func(t *testing.T) {
		prior := map[string]ClickHouseClusterModel{"old": cluster(map[string]ClickHouseUserModel{"app": user(types.StringValue("kept"))})}
		planned := map[string]ClickHouseClusterModel{"new": cluster(map[string]ClickHouseUserModel{"app": user(types.StringNull())})}
		if !ValidateClickHouseUserPasswords(root, prior, planned).HasError() {
			t.Error("expected an error: the user does not exist in the new cluster")
		}
	})

	t.Run("a password type change needs a new password", func(t *testing.T) {
		prior := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{"app": user(types.StringValue("kept"))})}
		changed := user(types.StringNull())
		changed.Password.Type = types.StringValue("DOUBLE_SHA1_HEX")
		planned := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{"app": changed})}
		if !ValidateClickHouseUserPasswords(root, prior, planned).HasError() {
			t.Error("expected an error: the API ignores a type change without enc_value")
		}
	})

	t.Run("a set or unknown enc_value passes", func(t *testing.T) {
		planned := map[string]ClickHouseClusterModel{"ch": cluster(map[string]ClickHouseUserModel{
			"app":   user(types.StringValue("x")),
			"other": user(types.StringUnknown()),
		})}
		if diags := ValidateClickHouseUserPasswords(root, nil, planned); diags.HasError() {
			t.Errorf("unexpected diags: %v", diags)
		}
	})
}
