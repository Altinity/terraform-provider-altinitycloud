package env

import (
	"context"
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testList(t *testing.T, vals ...string) types.List {
	t.Helper()
	l, diags := ListToModel(vals)
	if diags.HasError() {
		t.Fatalf("ListToModel: %v", diags)
	}
	return l
}

func minimalClusterModel(name string) ClickHouseClusterModel {
	return ClickHouseClusterModel{
		Name:         types.StringValue(name),
		Mode:         types.StringNull(),
		Image:        types.StringValue("altinity/clickhouse-server:24.8"),
		InstanceType: types.StringValue("m6i.large"),
		Zones:        types.ListNull(types.StringType),
		Shards:       types.Int64Value(1),
		Replicas:     types.Int64Value(2),
		Stopped:      types.BoolValue(false),
		Disk:         &ClickHouseDiskModel{Size: types.Int64Value(100)},
		Keeper:       &ClickHouseKeeperRefModel{Enabled: types.BoolValue(true), Name: types.StringValue("keeper")},
	}
}

func TestClickHouseClustersToSDK(t *testing.T) {
	ctx := context.Background()

	t.Run("no clusters returns nil", func(t *testing.T) {
		got, diags := ClickHouseClustersToSDK(ctx, nil)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("omitted optionals stay nil and the main disk is named default", func(t *testing.T) {
		got, diags := ClickHouseClustersToSDK(ctx, []ClickHouseClusterModel{minimalClusterModel("ch")})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 cluster, got %d", len(got))
		}

		c := got[0]
		if c.Mode != nil {
			t.Errorf("expected nil mode, got %v", *c.Mode)
		}
		if c.Zones != nil {
			t.Errorf("expected nil zones, got %v", c.Zones)
		}
		if c.Disk.Name != ClickHouseDefaultDiskName {
			t.Errorf("expected disk named %q, got %q", ClickHouseDefaultDiskName, c.Disk.Name)
		}
		if c.AdditionalDisks != nil || c.Settings != nil || c.Profiles != nil || c.Users != nil {
			t.Error("expected omitted list attributes to stay nil")
		}
		if c.Keeper == nil || !c.Keeper.Enabled || c.Keeper.Name != "keeper" {
			t.Errorf("unexpected keeper: %#v", c.Keeper)
		}
	})

	// Every Optional+Computed attribute without a schema default is unknown in the
	// plan, and a pointer method would turn that into an explicit "" or 0.
	t.Run("unknown attributes are omitted, never sent as empty values", func(t *testing.T) {
		model := minimalClusterModel("ch")
		model.Mode = types.StringUnknown()
		model.Zones = types.ListUnknown(types.StringType)
		model.Disk = &ClickHouseDiskModel{
			Size:         types.Int64Value(100),
			StorageClass: types.StringUnknown(),
			Iops:         types.Int64Unknown(),
			Throughput:   types.Int64Unknown(),
		}
		model.Users = []ClickHouseUserModel{{
			Name:    types.StringValue("app"),
			Profile: types.StringUnknown(),
			Quota:   types.StringUnknown(),
		}}

		got, diags := ClickHouseClustersToSDK(ctx, []ClickHouseClusterModel{model})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		c := got[0]
		if c.Mode != nil {
			t.Errorf("an unknown mode must be omitted, got %q", *c.Mode)
		}
		if c.Zones != nil {
			t.Errorf("unknown zones must be omitted, got %v", c.Zones)
		}
		if c.Disk.StorageClass != nil || c.Disk.Iops != nil || c.Disk.Throughput != nil {
			t.Errorf("unknown disk attributes must be omitted, got %#v", c.Disk)
		}
		if c.Users[0].Profile != nil || c.Users[0].Quota != nil {
			t.Errorf("unknown user attributes must be omitted, got %#v", c.Users[0])
		}
	})

	t.Run("unknown attributes are omitted from the update input too", func(t *testing.T) {
		model := minimalClusterModel("ch")
		model.Disk = &ClickHouseDiskModel{
			Size:       types.Int64Value(100),
			Iops:       types.Int64Unknown(),
			Throughput: types.Int64Unknown(),
		}

		got, diags := ClickHouseClustersToUpdateSDK(ctx, []ClickHouseClusterModel{model})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Disk.Iops != nil || got[0].Disk.Throughput != nil {
			t.Errorf("unknown disk attributes must be omitted, got %#v", got[0].Disk)
		}
	})

	t.Run("full cluster round trips every field", func(t *testing.T) {
		model := minimalClusterModel("ch")
		model.Mode = types.StringValue("SWARM")
		model.Zones = testList(t, "us-east-1a", "us-east-1b")
		model.Disk = &ClickHouseDiskModel{
			Size:         types.Int64Value(100),
			StorageClass: types.StringValue("gp3"),
			Iops:         types.Int64Value(3000),
			Throughput:   types.Int64Value(125),
		}
		model.AdditionalDisks = []ClickHouseAdditionalDiskModel{{
			Name: types.StringValue("disk1"),
			Size: types.Int64Value(50),
		}}
		model.Settings = []ClickHouseSettingModel{
			{Key: types.StringValue("max_concurrent_queries"), Value: types.StringValue("200")},
			{Key: types.StringValue("secret_setting"), ValueFromSecret: &ClickHouseSecretRefModel{
				Name: types.StringValue("my-secret"), Key: types.StringValue("my-key"),
			}},
		}
		model.Profiles = []ClickHouseProfileModel{{
			Name:     types.StringValue("readonly"),
			Settings: []ClickHouseSettingModel{{Key: types.StringValue("readonly"), Value: types.StringValue("1")}},
		}}
		model.Users = []ClickHouseUserModel{{
			Name:          types.StringValue("app"),
			Profile:       types.StringValue("readonly"),
			AllowedCIDRs:  testList(t, "10.0.0.0/8"),
			PasswordType:  types.StringValue("SHA256_HEX"),
			PasswordValue: types.StringValue("deadbeef"),
		}}

		got, diags := ClickHouseClustersToSDK(ctx, []ClickHouseClusterModel{model})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		c := got[0]
		if c.Mode == nil || *c.Mode != sdk.ClickHouseClusterModeSpecSwarm {
			t.Errorf("unexpected mode: %v", c.Mode)
		}
		if len(c.Zones) != 2 {
			t.Errorf("expected 2 zones, got %v", c.Zones)
		}
		if c.Disk.StorageClass == nil || *c.Disk.StorageClass != "gp3" {
			t.Errorf("unexpected storage class: %v", c.Disk.StorageClass)
		}
		if len(c.AdditionalDisks) != 1 || c.AdditionalDisks[0].Name != "disk1" {
			t.Errorf("unexpected additional disks: %#v", c.AdditionalDisks)
		}
		if c.Settings[0].Value == nil || *c.Settings[0].Value != "200" {
			t.Errorf("unexpected setting value: %v", c.Settings[0].Value)
		}
		if c.Settings[1].ValueFromSecret == nil || c.Settings[1].ValueFromSecret.Key != "my-key" {
			t.Errorf("unexpected setting secret: %#v", c.Settings[1].ValueFromSecret)
		}
		if len(c.Profiles) != 1 || len(c.Profiles[0].Settings) != 1 {
			t.Errorf("unexpected profiles: %#v", c.Profiles)
		}
		u := c.Users[0]
		if u.PasswordValue == nil || *u.PasswordValue != "deadbeef" {
			t.Errorf("unexpected password value: %v", u.PasswordValue)
		}
		if u.PasswordType == nil || *u.PasswordType != sdk.ClickHouseUserPasswordTypeSpecInputSha256Hex {
			t.Errorf("unexpected password type: %v", u.PasswordType)
		}
		if len(u.AllowedCIDRs) != 1 {
			t.Errorf("unexpected allowed cidrs: %v", u.AllowedCIDRs)
		}
	})
}

func TestClickHouseClustersToUpdateSDK(t *testing.T) {
	ctx := context.Background()
	model := minimalClusterModel("ch")
	model.Mode = types.StringValue("SWARM")
	model.Zones = testList(t, "us-east-1a")
	model.Disk = &ClickHouseDiskModel{Size: types.Int64Value(200), StorageClass: types.StringValue("gp3")}

	got, diags := ClickHouseClustersToUpdateSDK(ctx, []ClickHouseClusterModel{model})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	c := got[0]
	if c.Disk == nil || c.Disk.Size == nil || *c.Disk.Size != 200 {
		t.Errorf("unexpected disk: %#v", c.Disk)
	}
	if c.Disk.Name != ClickHouseDefaultDiskName {
		t.Errorf("expected disk named %q, got %q", ClickHouseDefaultDiskName, c.Disk.Name)
	}
	// The update input has no mode/zones/storageClass fields at all: they are immutable.
	if c.Image == nil || *c.Image != "altinity/clickhouse-server:24.8" {
		t.Errorf("unexpected image: %v", c.Image)
	}
}

func TestClickHouseKeepersToUpdateSDK(t *testing.T) {
	got := ClickHouseKeepersToUpdateSDK([]ClickHouseKeeperModel{{
		Name:         types.StringValue("keeper"),
		InstanceType: types.StringValue("m6i.large"),
		Zones:        testList(t, "us-east-1a"),
		HA:           types.BoolValue(true),
		Stopped:      types.BoolValue(false),
		Disk:         &ClickHouseDiskModel{Size: types.Int64Value(30)},
	}})

	if len(got) != 1 {
		t.Fatalf("expected 1 keeper, got %d", len(got))
	}
	if got[0].Ha == nil || !*got[0].Ha {
		t.Errorf("unexpected ha: %v", got[0].Ha)
	}
	if got[0].Disk == nil || got[0].Disk.Name != ClickHouseDefaultDiskName {
		t.Errorf("unexpected disk: %#v", got[0].Disk)
	}
}

func TestClickHouseNamesToDelete(t *testing.T) {
	prior := []ClickHouseClusterModel{minimalClusterModel("keep"), minimalClusterModel("drop")}
	planned := []ClickHouseClusterModel{minimalClusterModel("keep"), minimalClusterModel("new")}

	got := ClickHouseClusterNamesToDelete(prior, planned)
	if len(got) != 1 || got[0] != "drop" {
		t.Errorf("expected [drop], got %v", got)
	}

	if got := ClickHouseClusterNamesToDelete(nil, planned); got != nil {
		t.Errorf("expected nil with no prior state, got %v", got)
	}
	if got := ClickHouseClusterNamesToDelete(prior, prior); got != nil {
		t.Errorf("expected nil when nothing was dropped, got %v", got)
	}

	keepersPrior := []ClickHouseKeeperModel{{Name: types.StringValue("a")}, {Name: types.StringValue("b")}}
	keepersPlanned := []ClickHouseKeeperModel{{Name: types.StringValue("b")}}
	if got := ClickHouseKeeperNamesToDelete(keepersPrior, keepersPlanned); len(got) != 1 || got[0] != "a" {
		t.Errorf("expected [a], got %v", got)
	}
}

func TestApplyClickHouseClusterNestedDeletes(t *testing.T) {
	prior := []ClickHouseClusterModel{{
		Name:            types.StringValue("ch"),
		AdditionalDisks: []ClickHouseAdditionalDiskModel{{Name: types.StringValue("disk1")}, {Name: types.StringValue("disk2")}},
		Settings: []ClickHouseSettingModel{
			{Key: types.StringValue("kept")},
			{Key: types.StringValue("gone")},
		},
		Profiles: []ClickHouseProfileModel{
			{Name: types.StringValue("readonly"), Settings: []ClickHouseSettingModel{
				{Key: types.StringValue("readonly")},
				{Key: types.StringValue("stale")},
			}},
			{Name: types.StringValue("dropped")},
		},
		Users: []ClickHouseUserModel{{Name: types.StringValue("app")}, {Name: types.StringValue("old")}},
	}}

	updates := []*sdk.ClickHouseClusterUpdateSpecInput{{
		Name:            "ch",
		AdditionalDisks: []*sdk.ClickHouseDiskUpdateSpecInput{{Name: "disk1"}},
		Settings:        []*sdk.ClickHouseSettingSpecInput{{Key: "kept"}},
		Profiles: []*sdk.ClickHouseProfileUpdateSpecInput{{
			Name:     "readonly",
			Settings: []*sdk.ClickHouseSettingSpecInput{{Key: "readonly"}},
		}},
		Users: []*sdk.ClickHouseUserSpecInput{{Name: "app"}},
	}}

	ApplyClickHouseClusterNestedDeletes(updates, prior)

	u := updates[0]
	if len(u.AdditionalDisksToDelete) != 1 || u.AdditionalDisksToDelete[0] != "disk2" {
		t.Errorf("expected [disk2], got %v", u.AdditionalDisksToDelete)
	}
	if len(u.SettingsToDelete) != 1 || u.SettingsToDelete[0] != "gone" {
		t.Errorf("expected [gone], got %v", u.SettingsToDelete)
	}
	if len(u.ProfilesToDelete) != 1 || u.ProfilesToDelete[0] != "dropped" {
		t.Errorf("expected [dropped], got %v", u.ProfilesToDelete)
	}
	if len(u.UsersToDelete) != 1 || u.UsersToDelete[0] != "old" {
		t.Errorf("expected [old], got %v", u.UsersToDelete)
	}
	if len(u.Profiles[0].SettingsToDelete) != 1 || u.Profiles[0].SettingsToDelete[0] != "stale" {
		t.Errorf("expected [stale], got %v", u.Profiles[0].SettingsToDelete)
	}
}

func TestApplyClickHouseClusterNestedDeletesIgnoresNewClusters(t *testing.T) {
	updates := []*sdk.ClickHouseClusterUpdateSpecInput{{Name: "brand-new"}}
	ApplyClickHouseClusterNestedDeletes(updates, []ClickHouseClusterModel{{Name: types.StringValue("other")}})

	if updates[0].UsersToDelete != nil || updates[0].SettingsToDelete != nil {
		t.Errorf("expected no deletes for a cluster absent from prior state: %#v", updates[0])
	}
}

func minimalClusterSpec(name string) ClickHouseClusterSpec {
	return ClickHouseClusterSpec{
		Name:         name,
		Mode:         "STANDARD",
		Image:        "altinity/clickhouse-server:24.8",
		InstanceType: "m6i.large",
		Zones:        []string{"us-east-1a"},
		Shards:       1,
		Replicas:     2,
		Disk:         ClickHouseDiskSpec{Name: "default", Size: 100},
		KeeperName:   strPtr("keeper"),
	}
}

func TestClickHouseClustersToModel(t *testing.T) {
	t.Run("no clusters stays null, never an empty list", func(t *testing.T) {
		got, diags := ClickHouseClustersToModel(nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got != nil {
			t.Errorf("expected nil slice, got %#v", got)
		}
	})

	t.Run("clusters come back in config order", func(t *testing.T) {
		prior := []ClickHouseClusterModel{minimalClusterModel("b"), minimalClusterModel("a")}
		specs := []ClickHouseClusterSpec{minimalClusterSpec("a"), minimalClusterSpec("b")}

		got, diags := ClickHouseClustersToModel(prior, specs)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Name.ValueString() != "b" || got[1].Name.ValueString() != "a" {
			t.Errorf("expected [b a], got [%s %s]", got[0].Name, got[1].Name)
		}
	})

	t.Run("empty API lists map to null, not to empty lists", func(t *testing.T) {
		spec := minimalClusterSpec("ch")
		spec.Users = []ClickHouseUserSpec{{Name: "app", AllowedCIDRs: []string{}, Databases: []string{}}}

		got, diags := ClickHouseClustersToModel(nil, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].AdditionalDisks != nil || got[0].Settings != nil || got[0].Profiles != nil {
			t.Error("expected omitted nested lists to stay nil")
		}
		if !got[0].Users[0].AllowedCIDRs.IsNull() || !got[0].Users[0].Databases.IsNull() {
			t.Error("expected empty API string lists to map to null")
		}
	})

	t.Run("a setting read from a secret keeps a null value", func(t *testing.T) {
		spec := minimalClusterSpec("ch")
		spec.Settings = []ClickHouseSettingSpec{
			{Key: "plain", Value: "1"},
			{Key: "secret", Value: "", ValueFromSecret: &ClickHouseSecretRefSpec{Name: "s", Key: "k"}},
		}

		got, diags := ClickHouseClustersToModel(nil, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Settings[0].Value.ValueString() != "1" {
			t.Errorf("unexpected plain value: %s", got[0].Settings[0].Value)
		}
		if !got[0].Settings[1].Value.IsNull() {
			t.Errorf("expected null value for a secret-backed setting, got %s", got[0].Settings[1].Value)
		}
		if got[0].Settings[1].ValueFromSecret == nil {
			t.Error("expected the secret reference to survive")
		}
	})

	t.Run("passwords come from prior state, never from the API", func(t *testing.T) {
		prior := minimalClusterModel("ch")
		prior.Users = []ClickHouseUserModel{{
			Name:          types.StringValue("app"),
			PasswordType:  types.StringValue("SHA256_HEX"),
			PasswordValue: types.StringValue("deadbeef"),
		}}

		spec := minimalClusterSpec("ch")
		plainText := "PLAIN_TEXT"
		spec.Users = []ClickHouseUserSpec{{Name: "app", PasswordType: &plainText}}

		got, diags := ClickHouseClustersToModel([]ClickHouseClusterModel{prior}, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Users[0].PasswordValue.ValueString() != "deadbeef" {
			t.Errorf("expected the config password to survive, got %s", got[0].Users[0].PasswordValue)
		}
		if got[0].Users[0].PasswordType.ValueString() != "SHA256_HEX" {
			t.Errorf("expected the config password type to survive, got %s", got[0].Users[0].PasswordType)
		}
	})

	t.Run("a user absent from prior state has a null password", func(t *testing.T) {
		spec := minimalClusterSpec("ch")
		spec.Users = []ClickHouseUserSpec{{Name: "imported"}}

		got, diags := ClickHouseClustersToModel(nil, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if !got[0].Users[0].PasswordValue.IsNull() {
			t.Errorf("expected a null password, got %s", got[0].Users[0].PasswordValue)
		}
	})

	t.Run("a detached keeper reports disabled and keeps the configured name", func(t *testing.T) {
		prior := minimalClusterModel("ch")
		prior.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolValue(false), Name: types.StringValue("keeper")}

		spec := minimalClusterSpec("ch")
		spec.KeeperName = nil

		got, diags := ClickHouseClustersToModel([]ClickHouseClusterModel{prior}, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Keeper.Enabled.ValueBool() {
			t.Error("expected keeper to be reported as disabled")
		}
		if got[0].Keeper.Name.ValueString() != "keeper" {
			t.Errorf("expected the configured keeper name to survive, got %s", got[0].Keeper.Name)
		}
	})

	t.Run("nested lists follow config order", func(t *testing.T) {
		prior := minimalClusterModel("ch")
		prior.Settings = []ClickHouseSettingModel{{Key: types.StringValue("b")}, {Key: types.StringValue("a")}}
		prior.Users = []ClickHouseUserModel{{Name: types.StringValue("y")}, {Name: types.StringValue("x")}}
		prior.AdditionalDisks = []ClickHouseAdditionalDiskModel{{Name: types.StringValue("disk2")}, {Name: types.StringValue("disk1")}}

		spec := minimalClusterSpec("ch")
		spec.Settings = []ClickHouseSettingSpec{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}
		spec.Users = []ClickHouseUserSpec{{Name: "x"}, {Name: "y"}}
		spec.AdditionalDisks = []ClickHouseDiskSpec{{Name: "disk1"}, {Name: "disk2"}}

		got, diags := ClickHouseClustersToModel([]ClickHouseClusterModel{prior}, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Settings[0].Key.ValueString() != "b" {
			t.Errorf("expected settings in config order, got %s first", got[0].Settings[0].Key)
		}
		if got[0].Users[0].Name.ValueString() != "y" {
			t.Errorf("expected users in config order, got %s first", got[0].Users[0].Name)
		}
		if got[0].AdditionalDisks[0].Name.ValueString() != "disk2" {
			t.Errorf("expected disks in config order, got %s first", got[0].AdditionalDisks[0].Name)
		}
	})
}

func TestClickHouseKeepersToModel(t *testing.T) {
	t.Run("no keepers stays null", func(t *testing.T) {
		got, diags := ClickHouseKeepersToModel(nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got != nil {
			t.Errorf("expected nil slice, got %#v", got)
		}
	})

	t.Run("keepers come back in config order", func(t *testing.T) {
		prior := []ClickHouseKeeperModel{{Name: types.StringValue("b")}, {Name: types.StringValue("a")}}
		specs := []ClickHouseKeeperSpec{
			{Name: "a", Disk: ClickHouseDiskSpec{Size: 30}},
			{Name: "b", Disk: ClickHouseDiskSpec{Size: 30}},
		}

		got, diags := ClickHouseKeepersToModel(prior, specs)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got[0].Name.ValueString() != "b" || got[1].Name.ValueString() != "a" {
			t.Errorf("expected [b a], got [%s %s]", got[0].Name, got[1].Name)
		}
		if got[0].Disk == nil || got[0].Disk.Size.ValueInt64() != 30 {
			t.Errorf("unexpected disk: %#v", got[0].Disk)
		}
	})

	t.Run("zones follow config order", func(t *testing.T) {
		prior := []ClickHouseKeeperModel{{
			Name:  types.StringValue("keeper"),
			Zones: testList(t, "us-east-1b", "us-east-1a"),
		}}
		specs := []ClickHouseKeeperSpec{{Name: "keeper", Zones: []string{"us-east-1a", "us-east-1b"}}}

		got, diags := ClickHouseKeepersToModel(prior, specs)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		assertZones(t, got[0].Zones, "us-east-1b", "us-east-1a")
	})
}

func assertZones(t *testing.T, got types.List, want ...string) {
	t.Helper()
	elements := got.Elements()
	if len(elements) != len(want) {
		t.Fatalf("expected %d zones, got %v", len(want), got)
	}
	for i, w := range want {
		s, ok := elements[i].(types.String)
		if !ok {
			t.Fatalf("zone %d is not a types.String: %T", i, elements[i])
		}
		if s.ValueString() != w {
			t.Errorf("zone %d: got %q, want %q", i, s.ValueString(), w)
		}
	}
}

// The API returns zones in its own order, which would diff forever against an
// immutable list attribute.
func TestClickHouseZonesFollowConfigOrder(t *testing.T) {
	t.Run("cluster zones are reordered against prior state", func(t *testing.T) {
		prior := minimalClusterModel("ch")
		prior.Zones = testList(t, "us-east-1c", "us-east-1a", "us-east-1b")

		spec := minimalClusterSpec("ch")
		spec.Zones = []string{"us-east-1a", "us-east-1b", "us-east-1c"}

		got, diags := ClickHouseClustersToModel([]ClickHouseClusterModel{prior}, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		assertZones(t, got[0].Zones, "us-east-1c", "us-east-1a", "us-east-1b")
	})

	t.Run("zones the config never named go last", func(t *testing.T) {
		prior := minimalClusterModel("ch")
		prior.Zones = testList(t, "us-east-1b")

		spec := minimalClusterSpec("ch")
		spec.Zones = []string{"us-east-1a", "us-east-1b"}

		got, diags := ClickHouseClustersToModel([]ClickHouseClusterModel{prior}, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		assertZones(t, got[0].Zones, "us-east-1b", "us-east-1a")
	})

	t.Run("without prior state the API order is kept", func(t *testing.T) {
		spec := minimalClusterSpec("ch")
		spec.Zones = []string{"us-east-1a", "us-east-1b"}

		got, diags := ClickHouseClustersToModel(nil, []ClickHouseClusterSpec{spec})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		assertZones(t, got[0].Zones, "us-east-1a", "us-east-1b")
	})
}
