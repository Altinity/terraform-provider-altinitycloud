package env

import (
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func keeperModel() ClickHouseKeeperModel {
	return ClickHouseKeeperModel{
		Name:         types.StringValue("keeper"),
		InstanceType: types.StringValue("t4g.small"),
		Zones:        types.ListNull(types.StringType),
		HA:           types.BoolValue(true),
		Stopped:      types.BoolValue(false),
		Disk:         &ClickHouseDiskModel{Size: types.Int64Value(30)},
	}
}

func nodeGroup(t *testing.T, nodeType string, reservations ...sdk.NodeReservation) NodeGroupPlacement {
	t.Helper()
	set, diags := ReservationsToModel(reservations)
	if diags.HasError() {
		t.Fatalf("ReservationsToModel: %v", diags)
	}
	return NodeGroupPlacement{NodeType: types.StringValue(nodeType), Reservations: set}
}

func testNodeGroups(t *testing.T) []NodeGroupPlacement {
	t.Helper()
	return []NodeGroupPlacement{
		nodeGroup(t, "m6i.large", sdk.NodeReservationClickhouse),
		nodeGroup(t, "t4g.small", sdk.NodeReservationSystem, sdk.NodeReservationZookeeper),
	}
}

func TestValidateClickHouseConfig(t *testing.T) {
	t.Run("a cluster pointing at a declared keeper passes", func(t *testing.T) {
		diags := ValidateClickHouseConfig(
			[]ClickHouseClusterModel{minimalClusterModel("ch")},
			[]ClickHouseKeeperModel{keeperModel()},
			testNodeGroups(t),
		)
		if diags.HasError() {
			t.Fatalf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("an undeclared keeper is rejected", func(t *testing.T) {
		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{minimalClusterModel("ch")}, nil, nil)
		if !diags.HasError() {
			t.Fatal("expected an error for a keeper that is not declared")
		}
	})

	t.Run("only a SWARM cluster may run without a keeper", func(t *testing.T) {
		standard := minimalClusterModel("ch")
		standard.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolValue(false), Name: types.StringValue("")}
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{standard}, nil, nil); !diags.HasError() {
			t.Error("expected an error for a STANDARD cluster with the keeper disabled")
		}

		swarm := standard
		swarm.Mode = types.StringValue("SWARM")
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{swarm}, nil, nil); diags.HasError() {
			t.Errorf("unexpected errors for a SWARM cluster: %v", diags.Errors())
		}

		unknownMode := standard
		unknownMode.Mode = types.StringUnknown()
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{unknownMode}, nil, nil); diags.HasError() {
			t.Errorf("an unknown mode may still turn out to be SWARM: %v", diags.Errors())
		}
	})

	t.Run("a keeper name is required only while the keeper is enabled", func(t *testing.T) {
		missing := minimalClusterModel("ch")
		missing.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolValue(true), Name: types.StringNull()}
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{missing}, []ClickHouseKeeperModel{keeperModel()}, nil); !diags.HasError() {
			t.Error("expected an error for an enabled keeper with no name")
		}

		swarm := minimalClusterModel("ch")
		swarm.Mode = types.StringValue("SWARM")
		swarm.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolValue(false), Name: types.StringNull()}
		if diags := ValidateClickHouseConfig([]ClickHouseClusterModel{swarm}, nil, nil); diags.HasError() {
			t.Errorf("a SWARM cluster without a Keeper needs no name: %v", diags.Errors())
		}
	})

	// ValidateConfig runs before schema defaults, so an omitted `enabled` is null.
	t.Run("an omitted keeper.enabled is the default, not an opt-out", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		cluster.Keeper = &ClickHouseKeeperRefModel{Enabled: types.BoolNull(), Name: types.StringValue("keeper")}

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()}, nil)
		if diags.HasError() {
			t.Fatalf("a cluster that omits `enabled` still coordinates through a Keeper: %v", diags.Errors())
		}
	})

	t.Run("duplicate names are rejected", func(t *testing.T) {
		clusters := []ClickHouseClusterModel{minimalClusterModel("ch"), minimalClusterModel("ch")}
		if diags := ValidateClickHouseConfig(clusters, []ClickHouseKeeperModel{keeperModel()}, nil); !diags.HasError() {
			t.Error("expected an error for duplicate cluster names")
		}

		keepers := []ClickHouseKeeperModel{keeperModel(), keeperModel()}
		if diags := ValidateClickHouseConfig(nil, keepers, nil); !diags.HasError() {
			t.Error("expected an error for duplicate keeper names")
		}
	})

	t.Run("unknown names are not duplicates of each other", func(t *testing.T) {
		a := minimalClusterModel("a")
		a.Name = types.StringUnknown()
		b := minimalClusterModel("b")
		b.Name = types.StringUnknown()

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{a, b}, []ClickHouseKeeperModel{keeperModel()}, nil)
		if diags.HasError() {
			t.Errorf("two unknown names are not known to collide: %v", diags.Errors())
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

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()}, nil)
		if len(diags.Errors()) != 4 {
			t.Errorf("expected 4 duplicate errors, got %d: %v", len(diags.Errors()), diags.Errors())
		}
	})
}

func TestValidateClickHouseNodeGroupPlacement(t *testing.T) {
	t.Run("an instance type without a matching reservation is rejected", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		cluster.InstanceType = types.StringValue("t4g.small") // reserved for ZOOKEEPER, not CLICKHOUSE

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()}, testNodeGroups(t))
		if !diags.HasError() {
			t.Error("expected an error for a cluster on a node group without a CLICKHOUSE reservation")
		}
	})

	t.Run("an instance type absent from every node group is rejected", func(t *testing.T) {
		keeper := keeperModel()
		keeper.InstanceType = types.StringValue("c7g.xlarge")

		diags := ValidateClickHouseConfig(nil, []ClickHouseKeeperModel{keeper}, testNodeGroups(t))
		if !diags.HasError() {
			t.Error("expected an error for a Keeper with no matching node group")
		}
	})

	t.Run("the check is skipped while node groups are unsettled", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		cluster.InstanceType = types.StringValue("c7g.xlarge")
		unsettled := []NodeGroupPlacement{{NodeType: types.StringUnknown(), Reservations: types.SetNull(types.StringType)}}

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()}, unsettled)
		if diags.HasError() {
			t.Errorf("an unknown node group cannot rule out a match: %v", diags.Errors())
		}
	})

	t.Run("the check is skipped when no node groups were read", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		cluster.InstanceType = types.StringValue("c7g.xlarge")

		diags := ValidateClickHouseConfig([]ClickHouseClusterModel{cluster}, []ClickHouseKeeperModel{keeperModel()}, nil)
		if diags.HasError() {
			t.Errorf("unexpected errors: %v", diags.Errors())
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

	t.Run("an unknown zone element defers the comparison", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.Zones = testList(t, "us-east-1a", "us-east-1b")

		plan := state
		partial, diags := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("us-east-1a"),
			types.StringUnknown(),
		})
		if diags.HasError() {
			t.Fatalf("ListValue: %v", diags)
		}
		plan.Zones = partial

		if diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil); diags.HasError() {
			t.Errorf("an unknown element cannot be compared yet: %v", diags.Errors())
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

// The update input carries no mode, zones or storage class, so an entry added to
// an existing environment would silently come up with the defaults.
func TestValidateClickHousePlanEntriesAddedToExistingEnv(t *testing.T) {
	existing := minimalClusterModel("existing")

	t.Run("a plain new cluster is allowed", func(t *testing.T) {
		added := minimalClusterModel("added")
		added.Mode = types.StringUnknown()
		added.Zones = types.ListUnknown(types.StringType)
		added.Disk = &ClickHouseDiskModel{Size: types.Int64Value(100), StorageClass: types.StringUnknown()}

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{existing}, []ClickHouseClusterModel{existing, added}, nil, nil)
		if diags.HasError() {
			t.Fatalf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("an explicit STANDARD mode is allowed, since that is what would be created", func(t *testing.T) {
		added := minimalClusterModel("added")
		added.Mode = types.StringValue("STANDARD")

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{existing}, []ClickHouseClusterModel{existing, added}, nil, nil)
		if diags.HasError() {
			t.Errorf("unexpected errors: %v", diags.Errors())
		}
	})

	t.Run("mode, zones and storage_class are rejected", func(t *testing.T) {
		added := minimalClusterModel("added")
		added.Mode = types.StringValue("SWARM")
		added.Zones = testList(t, "us-east-1a")
		added.Disk = &ClickHouseDiskModel{Size: types.Int64Value(100), StorageClass: types.StringValue("gp3")}

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{existing}, []ClickHouseClusterModel{existing, added}, nil, nil)
		if len(diags.Errors()) != 3 {
			t.Errorf("expected 3 errors, got %d: %v", len(diags.Errors()), diags.Errors())
		}
	})

	t.Run("a volume added to an existing cluster cannot pick a storage class", func(t *testing.T) {
		plan := existing
		plan.AdditionalDisks = []ClickHouseAdditionalDiskModel{{
			Name:         types.StringValue("disk1"),
			Size:         types.Int64Value(50),
			StorageClass: types.StringValue("gp3"),
		}}

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{existing}, []ClickHouseClusterModel{plan}, nil, nil)
		if !diags.HasError() {
			t.Error("expected an error for a new volume with an explicit storage class")
		}
	})

	t.Run("a new keeper cannot pick zones or a storage class", func(t *testing.T) {
		added := keeperModel()
		added.Name = types.StringValue("added")
		added.Zones = testList(t, "us-east-1a")
		added.Disk = &ClickHouseDiskModel{Size: types.Int64Value(30), StorageClass: types.StringValue("gp3")}

		diags := ValidateClickHousePlan(nil, nil, []ClickHouseKeeperModel{keeperModel()}, []ClickHouseKeeperModel{keeperModel(), added})
		if len(diags.Errors()) != 2 {
			t.Errorf("expected 2 errors, got %d: %v", len(diags.Errors()), diags.Errors())
		}
	})
}

func TestValidateClickHousePlanWarnsOnDeletion(t *testing.T) {
	t.Run("dropping a cluster warns without blocking", func(t *testing.T) {
		diags := ValidateClickHousePlan(
			[]ClickHouseClusterModel{minimalClusterModel("gone")},
			nil, nil, nil,
		)
		if diags.HasError() {
			t.Fatalf("deletion is allowed: %v", diags.Errors())
		}
		if len(diags.Warnings()) != 1 {
			t.Errorf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
		}
	})

	t.Run("renaming warns about the entry left behind", func(t *testing.T) {
		diags := ValidateClickHousePlan(
			[]ClickHouseClusterModel{minimalClusterModel("old")},
			[]ClickHouseClusterModel{minimalClusterModel("new")},
			nil, nil,
		)
		if len(diags.Warnings()) != 1 {
			t.Errorf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
		}
	})

	t.Run("dropping a keeper warns", func(t *testing.T) {
		diags := ValidateClickHousePlan(nil, nil, []ClickHouseKeeperModel{keeperModel()}, nil)
		if len(diags.Warnings()) != 1 {
			t.Errorf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
		}
	})

	t.Run("dropping an additional volume warns", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.AdditionalDisks = []ClickHouseAdditionalDiskModel{{Name: types.StringValue("disk1"), Size: types.Int64Value(50)}}

		plan := state
		plan.AdditionalDisks = nil

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil)
		if diags.HasError() {
			t.Fatalf("deletion is allowed: %v", diags.Errors())
		}
		if len(diags.Warnings()) != 1 {
			t.Errorf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
		}
	})

	t.Run("renaming an additional volume warns about the data left behind", func(t *testing.T) {
		state := minimalClusterModel("ch")
		state.AdditionalDisks = []ClickHouseAdditionalDiskModel{{Name: types.StringValue("disk1"), Size: types.Int64Value(50)}}

		plan := state
		plan.AdditionalDisks = []ClickHouseAdditionalDiskModel{{Name: types.StringValue("disk2"), Size: types.Int64Value(50)}}

		diags := ValidateClickHousePlan([]ClickHouseClusterModel{state}, []ClickHouseClusterModel{plan}, nil, nil)
		if len(diags.Warnings()) != 1 {
			t.Errorf("expected 1 warning, got %d: %v", len(diags.Warnings()), diags.Warnings())
		}
	})

	t.Run("an unchanged plan warns about nothing", func(t *testing.T) {
		cluster := minimalClusterModel("ch")
		diags := ValidateClickHousePlan([]ClickHouseClusterModel{cluster}, []ClickHouseClusterModel{cluster}, nil, nil)
		if len(diags.Warnings()) != 0 {
			t.Errorf("expected no warnings, got %v", diags.Warnings())
		}
	})
}
