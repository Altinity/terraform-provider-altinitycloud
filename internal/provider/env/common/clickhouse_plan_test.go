package env

import (
	"context"
	"testing"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func clusterAttribute() rschema.ListNestedAttribute {
	return clientsupport.GetClickHouseClustersAttribute(false, true, false)
}

func clusterList(t *testing.T, models []ClickHouseClusterModel) types.List {
	t.Helper()
	elemType := clusterAttribute().NestedObject.Type()
	if models == nil {
		return types.ListNull(elemType)
	}
	list, diags := types.ListValueFrom(context.Background(), elemType, models)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return list
}

func objectAt(t *testing.T, list types.List, i int) types.Object {
	t.Helper()
	return objectValue(t, list.Elements()[i])
}

func objectValue(t *testing.T, v attr.Value) types.Object {
	t.Helper()
	obj, ok := v.(types.Object)
	if !ok {
		t.Fatalf("expected an object, got %T", v)
	}
	return obj
}

// What the framework hands ModifyPlan for a cluster whose computed values the config omits.
func plannedCluster(name string, image types.String) ClickHouseClusterModel {
	return ClickHouseClusterModel{
		Name:         types.StringValue(name),
		Image:        image,
		InstanceType: types.StringValue("m6i.large"),
		Zones:        types.ListUnknown(types.StringType),
		Shards:       types.Int64Unknown(),
		Replicas:     types.Int64Unknown(),
		Stopped:      types.BoolValue(false),
		Keeper:       types.StringNull(),
		Disk:         &ClickHouseDiskModel{SizeGiB: types.Int64Value(100), Iops: types.Int64Unknown(), ThroughputMiBs: types.Int64Unknown()},
	}
}

func configuredCluster(name string, image types.String) ClickHouseClusterModel {
	c := plannedCluster(name, image)
	c.Zones = types.ListNull(types.StringType)
	c.Shards = types.Int64Null()
	c.Replicas = types.Int64Null()
	c.Disk = &ClickHouseDiskModel{SizeGiB: types.Int64Value(100), Iops: types.Int64Null(), ThroughputMiBs: types.Int64Null()}
	return c
}

func stateCluster(name, image string) ClickHouseClusterModel {
	spec := minimalClusterSpec(name)
	spec.Image = image
	return clickHouseClusterToModel(ClickHouseClusterModel{}, spec)
}

func fillClusters(t *testing.T, plan, state, config []ClickHouseClusterModel) []ClickHouseClusterModel {
	t.Helper()
	ctx := context.Background()
	got := UseStateForUnknownByKeyList(ctx, "clickhouse_clusters", clusterAttribute(),
		clusterList(t, plan), clusterList(t, state), clusterList(t, config))
	models, diags := ObjectsFromList[ClickHouseClusterModel](ctx, got)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return models
}

func TestUseStateForUnknownByKey(t *testing.T) {
	t.Run("an unrelated change keeps the computed values of an existing cluster", func(t *testing.T) {
		got := fillClusters(t,
			[]ClickHouseClusterModel{plannedCluster("a", types.StringUnknown())},
			[]ClickHouseClusterModel{stateCluster("a", "image-a")},
			[]ClickHouseClusterModel{configuredCluster("a", types.StringNull())},
		)

		c := got[0]
		if c.Image.ValueString() != "image-a" || c.Shards.ValueInt64() != 1 {
			t.Errorf("expected state values, got %#v", c)
		}
		// keeper is not computed: omitting it detaches the Keeper instead of keeping the state value.
		if !c.Keeper.IsNull() {
			t.Errorf("expected keeper to stay null, got %v", c.Keeper)
		}
		if c.Disk.Iops.ValueInt64() != 3000 || len(c.Zones.Elements()) != 1 {
			t.Errorf("nested computed values must come from state too, got %#v", c)
		}
	})

	t.Run("a renamed cluster does not inherit its predecessor's values", func(t *testing.T) {
		// Same length: Terraform pairs by index and hands b the image a had.
		got := fillClusters(t,
			[]ClickHouseClusterModel{plannedCluster("b", types.StringValue("image-a"))},
			[]ClickHouseClusterModel{stateCluster("a", "image-a")},
			[]ClickHouseClusterModel{configuredCluster("b", types.StringNull())},
		)

		if !got[0].Image.IsUnknown() || !got[0].Shards.IsUnknown() || !got[0].Disk.Iops.IsUnknown() {
			t.Errorf("a new entry must let the platform pick defaults, got %#v", got[0])
		}
	})

	t.Run("a reorder pairs by name", func(t *testing.T) {
		got := fillClusters(t,
			[]ClickHouseClusterModel{plannedCluster("b", types.StringValue("image-a")), plannedCluster("a", types.StringValue("image-b"))},
			[]ClickHouseClusterModel{stateCluster("a", "image-a"), stateCluster("b", "image-b")},
			[]ClickHouseClusterModel{configuredCluster("b", types.StringNull()), configuredCluster("a", types.StringNull())},
		)

		if got[0].Image.ValueString() != "image-b" || got[1].Image.ValueString() != "image-a" {
			t.Errorf("expected each cluster to keep its own image, got %v / %v", got[0].Image, got[1].Image)
		}
	})

	t.Run("configured and config-unknown values are left alone", func(t *testing.T) {
		got := fillClusters(t,
			[]ClickHouseClusterModel{plannedCluster("a", types.StringValue("custom")), plannedCluster("b", types.StringUnknown())},
			[]ClickHouseClusterModel{stateCluster("a", "image-a"), stateCluster("b", "image-b")},
			[]ClickHouseClusterModel{configuredCluster("a", types.StringValue("custom")), configuredCluster("b", types.StringUnknown())},
		)

		if got[0].Image.ValueString() != "custom" {
			t.Errorf("a configured value must win over state, got %v", got[0].Image)
		}
		if !got[1].Image.IsUnknown() {
			t.Errorf("a value unknown in the config must stay unknown, got %v", got[1].Image)
		}
	})

	t.Run("an attribute with a default keeps the default", func(t *testing.T) {
		state := stateCluster("a", "image-a")
		state.Stopped = types.BoolValue(true)
		got := fillClusters(t,
			[]ClickHouseClusterModel{plannedCluster("a", types.StringUnknown())},
			[]ClickHouseClusterModel{state},
			[]ClickHouseClusterModel{configuredCluster("a", types.StringNull())},
		)

		if got[0].Stopped.ValueBool() {
			t.Errorf("stopped must plan as its default, not the drifted state value")
		}
	})

	t.Run("nested entries pair by name too", func(t *testing.T) {
		user := func(name string, profile types.String) ClickHouseUserModel {
			return ClickHouseUserModel{
				Name:      types.StringValue(name),
				Profile:   profile,
				Databases: types.ListNull(types.StringType),
				Grants:    types.ListNull(types.StringType),
				Password:  &ClickHouseUserPasswordModel{Type: types.StringValue("SHA256_HEX"), EncValue: types.StringNull()},
			}
		}
		state := stateCluster("a", "image-a")
		state.Users = []ClickHouseUserModel{userToModel(ClickHouseUserModel{}, ClickHouseUserSpec{Name: "app", Profile: "readonly", PasswordType: "SHA256_HEX"})}
		plan := plannedCluster("a", types.StringUnknown())
		plan.Users = []ClickHouseUserModel{user("new", types.StringUnknown()), user("app", types.StringUnknown())}
		config := configuredCluster("a", types.StringNull())
		config.Users = []ClickHouseUserModel{user("new", types.StringNull()), user("app", types.StringNull())}

		got := fillClusters(t, []ClickHouseClusterModel{plan}, []ClickHouseClusterModel{state}, []ClickHouseClusterModel{config})

		users := got[0].Users
		if !users[0].Profile.IsUnknown() {
			t.Errorf("a new user must not inherit app's profile, got %v", users[0].Profile)
		}
		if users[1].Profile.ValueString() != "readonly" {
			t.Errorf("an existing user keeps its profile, got %v", users[1].Profile)
		}
	})

	t.Run("an unknown nested object is left as is", func(t *testing.T) {
		ctx := context.Background()
		obj := objectAt(t, clusterList(t, []ClickHouseClusterModel{plannedCluster("a", types.StringUnknown())}), 0)
		attrs := obj.Attributes()
		attrs["disk"] = types.ObjectUnknown(objectValue(t, attrs["disk"]).AttributeTypes(ctx))
		plan := types.ListValueMust(obj.Type(ctx), []attr.Value{types.ObjectValueMust(obj.AttributeTypes(ctx), attrs)})

		got := UseStateForUnknownByKeyList(ctx, "clickhouse_clusters", clusterAttribute(),
			plan, clusterList(t, []ClickHouseClusterModel{stateCluster("a", "image-a")}), clusterList(t, nil))

		if disk := objectAt(t, got, 0).Attributes()["disk"]; !disk.IsUnknown() {
			t.Errorf("expected disk to stay unknown, got %v", disk)
		}

		models, diags := ObjectsFromList[ClickHouseClusterModel](ctx, got)
		if diags.HasError() || models[0].Disk != nil {
			t.Errorf("a lenient conversion must tolerate the unknown object, got %#v / %v", models, diags)
		}
	})
}

func TestValidateClickHouseNewUserPasswords(t *testing.T) {
	user := func(name string, encValue types.String) ClickHouseUserModel {
		return ClickHouseUserModel{
			Name:     types.StringValue(name),
			Password: &ClickHouseUserPasswordModel{Type: types.StringValue("SHA256_HEX"), EncValue: encValue},
		}
	}
	cluster := func(name string, users ...ClickHouseUserModel) ClickHouseClusterModel {
		return ClickHouseClusterModel{Name: types.StringValue(name), Users: users}
	}
	root := path.Root("clickhouse_clusters")

	t.Run("a new user needs a password", func(t *testing.T) {
		diags := ValidateClickHouseNewUserPasswords(root, nil, []ClickHouseClusterModel{cluster("ch", user("app", types.StringNull()))})
		if !diags.HasError() {
			t.Fatal("expected an error for a new user without enc_value")
		}
		withPath, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
		if !ok {
			t.Fatalf("expected an attribute error, got %T", diags.Errors()[0])
		}
		want := root.AtListIndex(0).AtName("users").AtListIndex(0).AtName("password").AtName("enc_value")
		if got := withPath.Path(); !got.Equal(want) {
			t.Errorf("error path = %s, want %s", got, want)
		}
	})

	t.Run("an existing user may omit it", func(t *testing.T) {
		prior := []ClickHouseClusterModel{cluster("ch", user("app", types.StringValue("kept")))}
		diags := ValidateClickHouseNewUserPasswords(root, prior, []ClickHouseClusterModel{cluster("ch", user("app", types.StringNull()))})
		if diags.HasError() {
			t.Errorf("unexpected diags: %v", diags)
		}
	})

	t.Run("the same user name in a new cluster is a new user", func(t *testing.T) {
		prior := []ClickHouseClusterModel{cluster("old", user("app", types.StringValue("kept")))}
		diags := ValidateClickHouseNewUserPasswords(root, prior, []ClickHouseClusterModel{cluster("new", user("app", types.StringNull()))})
		if !diags.HasError() {
			t.Error("expected an error: the user does not exist in the new cluster")
		}
	})

	t.Run("a set or unknown enc_value passes", func(t *testing.T) {
		planned := []ClickHouseClusterModel{cluster("ch", user("app", types.StringValue("x")), user("other", types.StringUnknown()))}
		if diags := ValidateClickHouseNewUserPasswords(root, nil, planned); diags.HasError() {
			t.Errorf("unexpected diags: %v", diags)
		}
	})
}
