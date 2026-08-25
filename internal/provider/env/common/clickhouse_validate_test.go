package env

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func keeperModel() ClickHouseKeeperModel {
	return ClickHouseKeeperModel{
		Name:         types.StringValue("keeper"),
		InstanceType: types.StringValue("m6i.large"),
		Zones:        types.ListNull(types.StringType),
		HA:           types.BoolValue(true),
		Stopped:      types.BoolValue(false),
		Disk:         &ClickHouseDiskModel{Size: types.Int64Value(30)},
	}
}

func TestValidateClickHouseConfig(t *testing.T) {
	t.Run("a cluster pointing at a declared keeper passes", func(t *testing.T) {
		diags := ValidateClickHouseConfig(
			[]ClickHouseClusterModel{minimalClusterModel("ch")},
			[]ClickHouseKeeperModel{keeperModel()},
		)
		if diags.HasError() {
			t.Fatalf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("an undeclared keeper is rejected", func(t *testing.T) {
		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{minimalClusterModel("ch")}, nil)
		if !diags.HasError() {
			t.Fatal("expected an error for a keeper that is not declared")
		}
	})

	t.Run("only a SWARM cluster may run without a keeper", func(t *testing.T) {
		standard := minimalClusterModel("ch")
		standard.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolValue(false), Name: types.StringValue("")}
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{standard}, nil); !diags.HasError() {
			t.Error("expected an error for a STANDARD cluster with the keeper disabled")
		}

		swarm := standard
		swarm.Mode = types.StringValue("SWARM")
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{swarm}, nil); diags.HasError() {
			t.Errorf("unexpected errors for a SWARM cluster: %v", diags.Errors())
		}
	})

	t.Run("duplicate names are rejected", func(t *testing.T) {
		clusters := []ClickHouseClusterModel{minimalClusterModel("ch"), minimalClusterModel("ch")}
		if diags := ValidateClickHouseConfig(clusters, []ClickHouseKeeperModel{keeperModel()}); !diags.HasError() {
			t.Error("expected an error for duplicate cluster names")
		}

		keepers := []ClickHouseKeeperModel{keeperModel(), keeperModel()}
		if diags := ValidateClickHouseConfig(nil, keepers); !diags.HasError() {
			t.Error("expected an error for duplicate keeper names")
		}
	})

	t.Run("duplicate nested entries are rejected", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		cluster.Settings = []ClickHouseSettingModel{
			{Key: types.StringValue("max_concurrent_queries")},
			{Key: types.StringValue("max_concurrent_queries")},
		}
		cluster.Users = []ClickHouseUserModel{{Name: types.StringValue("app")}, {Name: types.StringValue("app")}}
		cluster.AdditionalDisks = []ClickHouseAdditionalDiskModel{
			{Name: types.StringValue("disk1")},
			{Name: types.StringValue("disk1")},
		}
		cluster.Profiles = []ClickHouseProfileModel{{
			Name: types.StringValue("readonly"),
			Settings: []ClickHouseSettingModel{
				{Key: types.StringValue("readonly")},
				{Key: types.StringValue("readonly")},
			},
		}}

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()})
		if len(diags.Errors()) != 4 {
			t.Errorf("expected 4 duplicate errors, got %d: %v", len(diags.Errors()), diags.Errors())
		}
	})
}

func TestValidateClickHousePlan(t *testing.T) {
	t.Run("an unchanged plan passes", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{cluster}, []ClickHouseClusterModel{cluster}, nil, nil); diags.HasError() {
			t.Fatalf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("changing an immutable attribute is rejected", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.Mode = types.StringValue("STANDARD")
		plan := state
		plan.Mode = types.StringValue("SWARM")

		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil); !diags.HasError() {
			t.Error("expected an error when mode changes")
		}
	})

	t.Run("zones are compared as a set", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.Zones = testList(t, "us-east-1a", "us-east-1b")

		reordered := state
		reordered.Zones = testList(t, "us-east-1b", "us-east-1a")
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{reordered}, nil, nil); diags.HasError() {
			t.Errorf("a pure reorder is not a change: %v", diags.Errors())
		}

		changed := state
		changed.Zones = testList(t, "us-east-1a", "us-east-1c")
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{changed}, nil, nil); !diags.HasError() {
			t.Error("expected an error when zones change")
		}
	})

	t.Run("volumes may grow but never shrink", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.Disk = &ClickHouseDiskModel{Size: types.Int64Value(100), StorageClass: types.StringValue("gp3")}

		grown := state
		grown.Disk = &ClickHouseDiskModel{Size: types.Int64Value(200), StorageClass: types.StringValue("gp3")}
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{grown}, nil, nil); diags.HasError() {
			t.Errorf("growing a volume is allowed: %v", diags.Errors())
		}

		shrunk := state
		shrunk.Disk = &ClickHouseDiskModel{Size: types.Int64Value(50), StorageClass: types.StringValue("gp3")}
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{shrunk}, nil, nil); !diags.HasError() {
			t.Error("expected an error when a volume shrinks")
		}

		reclassed := state
		reclassed.Disk = &ClickHouseDiskModel{Size: types.Int64Value(100), StorageClass: types.StringValue("io2")}
		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{reclassed}, nil, nil); !diags.HasError() {
			t.Error("expected an error when a storage class changes")
		}
	})

	t.Run("additional volumes are paired by name, not by position", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.AdditionalDisks = []ClickHouseAdditionalDiskModel{
			{Name: types.StringValue("disk1"), Size: types.Int64Value(50)},
			{Name: types.StringValue("disk2"), Size: types.Int64Value(200)},
		}

		plan := state
		plan.AdditionalDisks = []ClickHouseAdditionalDiskModel{
			{Name: types.StringValue("disk2"), Size: types.Int64Value(200)},
			{Name: types.StringValue("disk1"), Size: types.Int64Value(50)},
		}

		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil); diags.HasError() {
			t.Errorf("reordering volumes is not a shrink: %v", diags.Errors())
		}
	})

	t.Run("removing a cluster does not shift the comparison onto its neighbour", func(t *testing.T) {
		a := minimalClusterModel("a")
		a.Mode = types.StringValue("STANDARD")
		b := minimalClusterModel("b")
		b.Mode = types.StringValue("SWARM")

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{a, b}, []ClickHouseClusterModel{b}, nil, nil)
		if diags.HasError() {
			t.Errorf("dropping the first cluster is not a mode change: %v", diags.Errors())
		}
	})

	t.Run("a new cluster has nothing to compare against", func(t *testing.T) {
		if diags := ValidateClickHousePlan(nil, []ClickHouseClusterModel{minimalClusterModel("ch")}, nil, nil); diags.HasError() {
			t.Errorf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("an HA keeper cannot be shrunk back to a single node", func(t *testing.T) {
		state := keeperModel()
		plan := state
		plan.HA = types.BoolValue(false)
		if diags := ValidateClickHousePlan(nil, nil, []ClickHouseKeeperModel{state}, []ClickHouseKeeperModel{plan}); !diags.HasError() {
			t.Error("expected an error when ha is turned off")
		}

		single := keeperModel()
		single.HA = types.BoolValue(false)
		upgraded := keeperModel()
		if diags := ValidateClickHousePlan(nil, nil, []ClickHouseKeeperModel{single}, []ClickHouseKeeperModel{upgraded}); diags.HasError() {
			t.Errorf("turning ha on is allowed: %v", diags.Errors())
		}
	})

	t.Run("unknown plan values are skipped", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.Mode = types.StringValue("STANDARD")
		plan := state
		plan.Mode = types.StringUnknown()
		plan.Disk = &ClickHouseDiskModel{Size: types.Int64Unknown()}

		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil); diags.HasError() {
			t.Errorf("unknown values cannot be compared yet: %v", diags.Errors())
		}
	})
}
