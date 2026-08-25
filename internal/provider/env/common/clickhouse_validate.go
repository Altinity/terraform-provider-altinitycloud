package env

import (
	"context"
	"fmt"
	"sort"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var clickHouseClustersPath = path.Root("clickhouse_clusters")
var clickHouseKeepersPath = path.Root("clickhouse_keepers")
var nodeGroupsPath = path.Root("node_groups")

// Cross-attribute rules the schema validators cannot express on their own.
func ValidateClickHouseConfig(clusters []ClickHouseClusterModel, keepers []ClickHouseKeeperModel, nodeGroups []NodeGroupsModel) diag.Diagnostics {
	var diags diag.Diagnostics

	keeperNames := make(map[string]bool, len(keepers))
	for i, k := range keepers {
		keeperPath := clickHouseKeepersPath.AtListIndex(i)

		if name, ok := knownString(k.Name); ok {
			if keeperNames[name] {
				diags.AddAttributeError(keeperPath.AtName("name"), "Duplicate Keeper", fmt.Sprintf("Keeper %q is declared more than once.", name))
			}
			keeperNames[name] = true
		}

		diags.Append(validateNodeGroupPlacement(keeperPath, "Keeper", k.Name, k.InstanceType, sdk.NodeReservationZookeeper, nodeGroups)...)
	}

	clusterNames := make(map[string]bool, len(clusters))
	for i, c := range clusters {
		clusterPath := clickHouseClustersPath.AtListIndex(i)

		if name, ok := knownString(c.Name); ok {
			if clusterNames[name] {
				diags.AddAttributeError(clusterPath.AtName("name"), "Duplicate Cluster", fmt.Sprintf("ClickHouse cluster %q is declared more than once.", name))
			}
			clusterNames[name] = true
		}

		diags.Append(validateNodeGroupPlacement(clusterPath, "Cluster", c.Name, c.InstanceType, sdk.NodeReservationClickhouse, nodeGroups)...)
		diags.Append(validateClickHouseKeeperRef(clusterPath, c, keeperNames)...)
		diags.Append(validateUniqueNames(clusterPath.AtName("additional_disks"), "volume", c.AdditionalDisks,
			func(d ClickHouseAdditionalDiskModel) types.String { return d.Name })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("settings"), "setting", c.Settings,
			func(s ClickHouseSettingModel) types.String { return s.Key })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("profiles"), "profile", c.Profiles,
			func(p ClickHouseProfileModel) types.String { return p.Name })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("users"), "user", c.Users,
			func(u ClickHouseUserModel) types.String { return u.Name })...)

		for j, p := range c.Profiles {
			diags.Append(validateUniqueNames(clusterPath.AtName("profiles").AtListIndex(j).AtName("settings"), "setting", p.Settings,
				func(s ClickHouseSettingModel) types.String { return s.Key })...)
		}
	}

	return diags
}

func validateClickHouseKeeperRef(clusterPath path.Path, cluster ClickHouseClusterModel, keeperNames map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics
	if cluster.Keeper == nil {
		return diags
	}

	keeperPath := clusterPath.AtName("keeper")

	// ValidateConfig reads the raw config, before schema defaults are applied, so a
	// null `enabled` is the Default(true) and not an opt-out.
	settled := !cluster.Keeper.Enabled.IsNull() && !cluster.Keeper.Enabled.IsUnknown()
	if settled && !cluster.Keeper.Enabled.ValueBool() {
		// An unknown mode may still turn out to be SWARM; a null one is STANDARD.
		if !cluster.Mode.IsUnknown() && cluster.Mode.ValueString() != string(sdk.ClickHouseClusterModeSpecSwarm) {
			diags.AddAttributeError(keeperPath.AtName("enabled"), "Keeper Required",
				fmt.Sprintf("Cluster %q must coordinate through a Keeper. Only a SWARM cluster may set `enabled` to `false`.", cluster.Name.ValueString()))
		}
		return diags
	}

	name, ok := knownString(cluster.Keeper.Name)
	if !ok {
		return diags
	}

	if !keeperNames[name] {
		diags.AddAttributeError(keeperPath.AtName("name"), "Unknown Keeper",
			fmt.Sprintf("Cluster %q references Keeper %q, which is not declared in `clickhouse_keepers`.", cluster.Name.ValueString(), name))
	}

	return diags
}

// A cluster runs on nodes reserved for CLICKHOUSE and a Keeper on nodes reserved
// for ZOOKEEPER, which is knowable from the configuration alone.
func validateNodeGroupPlacement(attrPath path.Path, kind string, name, instanceType types.String, reservation sdk.NodeReservation, nodeGroups []NodeGroupsModel) diag.Diagnostics {
	var diags diag.Diagnostics

	wanted, ok := knownString(instanceType)
	if !ok || len(nodeGroups) == 0 {
		return diags
	}

	accepts, settled := nodeGroupAccepts(wanted, string(reservation), nodeGroups)
	if settled && !accepts {
		diags.AddAttributeError(attrPath.AtName("instance_type"), "No Matching Node Group",
			fmt.Sprintf("%s %q runs on %q, but no entry in `%s` uses that node type with a %s reservation.",
				kind, name.ValueString(), wanted, nodeGroupsPath, reservation))
	}

	return diags
}

// settled is false while any node group is still unknown, since a missing match
// would then be a guess.
func nodeGroupAccepts(instanceType, reservation string, nodeGroups []NodeGroupsModel) (accepts bool, settled bool) {
	settled = true

	for _, ng := range nodeGroups {
		if ng.NodeType.IsUnknown() || ng.Reservations.IsUnknown() || ng.Reservations.IsNull() {
			settled = false
			continue
		}
		if ng.NodeType.ValueString() != instanceType {
			continue
		}

		for _, element := range ng.Reservations.Elements() {
			s, ok := element.(types.String)
			if !ok || s.IsUnknown() || s.IsNull() {
				settled = false
				continue
			}
			if s.ValueString() == reservation {
				return true, true
			}
		}
	}

	return false, settled
}

func validateUniqueNames[T any](attrPath path.Path, kind string, items []T, key func(T) types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	seen := make(map[string]bool, len(items))

	for i, item := range items {
		k, ok := knownString(key(item))
		if !ok {
			continue
		}
		if seen[k] {
			diags.AddAttributeError(attrPath.AtListIndex(i), "Duplicate "+kind, fmt.Sprintf("%s %q is declared more than once.", kind, k))
		}
		seen[k] = true
	}

	return diags
}

// Entries are matched by name, so a config change the API rejects as immutable is
// caught here rather than halfway through an environment update.
func ValidateClickHousePlan(stateClusters, planClusters []ClickHouseClusterModel, stateKeepers, planKeepers []ClickHouseKeeperModel) diag.Diagnostics {
	var diags diag.Diagnostics

	priorClusters := make(map[string]ClickHouseClusterModel, len(stateClusters))
	for _, c := range stateClusters {
		priorClusters[c.Name.ValueString()] = c
	}

	plannedClusters := make(map[string]bool, len(planClusters))
	for i, c := range planClusters {
		clusterPath := clickHouseClustersPath.AtListIndex(i)
		plannedClusters[c.Name.ValueString()] = true

		prior, ok := priorClusters[c.Name.ValueString()]
		if !ok {
			diags.Append(validateClusterAddedToExistingEnv(clusterPath, c)...)
			continue
		}

		diags.Append(immutableString(clusterPath.AtName("mode"), "mode", prior.Mode, c.Mode)...)
		diags.Append(immutableList(clusterPath.AtName("zones"), "zones", prior.Zones, c.Zones)...)
		diags.Append(validateClickHouseDisk(clusterPath.AtName("disk"), prior.Disk, c.Disk)...)

		priorDisks := make(map[string]ClickHouseAdditionalDiskModel, len(prior.AdditionalDisks))
		for _, d := range prior.AdditionalDisks {
			priorDisks[d.Name.ValueString()] = d
		}
		for j, d := range c.AdditionalDisks {
			diskPath := clusterPath.AtName("additional_disks").AtListIndex(j)
			priorDisk, ok := priorDisks[d.Name.ValueString()]
			if !ok {
				diags.Append(setOnlyAtEnvCreation(diskPath.AtName("storage_class"), "storage_class", d.StorageClass, "volume")...)
				continue
			}
			diags.Append(immutableString(diskPath.AtName("storage_class"), "storage_class", priorDisk.StorageClass, d.StorageClass)...)
			diags.Append(growOnlyInt64(diskPath.AtName("size"), "size", priorDisk.Size, d.Size)...)
		}
	}

	priorKeepers := make(map[string]ClickHouseKeeperModel, len(stateKeepers))
	for _, k := range stateKeepers {
		priorKeepers[k.Name.ValueString()] = k
	}

	plannedKeepers := make(map[string]bool, len(planKeepers))
	for i, k := range planKeepers {
		keeperPath := clickHouseKeepersPath.AtListIndex(i)
		plannedKeepers[k.Name.ValueString()] = true

		prior, ok := priorKeepers[k.Name.ValueString()]
		if !ok {
			diags.Append(setOnlyAtEnvCreation(keeperPath.AtName("zones"), "zones", listPresence(k.Zones), "Keeper")...)
			if k.Disk != nil {
				diags.Append(setOnlyAtEnvCreation(keeperPath.AtName("disk").AtName("storage_class"), "storage_class", k.Disk.StorageClass, "Keeper")...)
			}
			continue
		}

		diags.Append(immutableList(keeperPath.AtName("zones"), "zones", prior.Zones, k.Zones)...)
		diags.Append(validateClickHouseDisk(keeperPath.AtName("disk"), prior.Disk, k.Disk)...)

		if prior.HA.ValueBool() && !k.HA.IsUnknown() && !k.HA.ValueBool() {
			diags.AddAttributeError(keeperPath.AtName("ha"), "Immutable Attribute",
				fmt.Sprintf("Keeper %q is running as a highly-available ensemble and cannot be shrunk back to a single node.", k.Name.ValueString()))
		}
	}

	diags.Append(warnClickHouseDeletions(clickHouseClustersPath, "ClickHouse cluster", stateClusters, plannedClusters,
		func(c ClickHouseClusterModel) types.String { return c.Name })...)
	diags.Append(warnClickHouseDeletions(clickHouseKeepersPath, "Keeper", stateKeepers, plannedKeepers,
		func(k ClickHouseKeeperModel) types.String { return k.Name })...)

	return diags
}

// The update API creates entries it cannot find by name, but its input carries no
// mode, zones or storage class, so those would silently fall back to defaults.
func validateClusterAddedToExistingEnv(clusterPath path.Path, cluster ClickHouseClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// STANDARD is what the update API would create anyway, so asking for it is fine.
	if mode, ok := knownString(cluster.Mode); ok && mode != string(sdk.ClickHouseClusterModeSpecStandard) {
		diags.Append(setOnlyAtEnvCreation(clusterPath.AtName("mode"), "mode", cluster.Mode, "cluster")...)
	}
	diags.Append(setOnlyAtEnvCreation(clusterPath.AtName("zones"), "zones", listPresence(cluster.Zones), "cluster")...)

	if cluster.Disk != nil {
		diags.Append(setOnlyAtEnvCreation(clusterPath.AtName("disk").AtName("storage_class"), "storage_class", cluster.Disk.StorageClass, "cluster")...)
	}
	for j, d := range cluster.AdditionalDisks {
		diags.Append(setOnlyAtEnvCreation(clusterPath.AtName("additional_disks").AtListIndex(j).AtName("storage_class"), "storage_class", d.StorageClass, "volume")...)
	}

	return diags
}

func setOnlyAtEnvCreation(attrPath path.Path, name string, value types.String, kind string) diag.Diagnostics {
	var diags diag.Diagnostics
	if _, ok := knownString(value); !ok {
		return diags
	}

	diags.AddAttributeError(attrPath, "Attribute Not Settable On Update",
		fmt.Sprintf("%s can only be set when the environment is created. A %s added to an existing environment goes through the update API, whose input carries no %s, so it would come up with the default.", name, kind, name))
	return diags
}

// Bridges a list attribute into the string-shaped presence check above: only
// whether the user set it matters.
func listPresence(list types.List) types.String {
	if list.IsNull() || list.IsUnknown() {
		return types.StringNull()
	}
	return types.StringValue("set")
}

func warnClickHouseDeletions[T any](attrPath path.Path, kind string, prior []T, planned map[string]bool, key func(T) types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, item := range prior {
		name, ok := knownString(key(item))
		if !ok || planned[name] {
			continue
		}

		diags.AddAttributeWarning(attrPath, kind+" Will Be Deleted",
			fmt.Sprintf("%s %q is no longer in the configuration and will be deleted along with its data. Note that renaming an entry is not a rename: the old one is deleted and a new, empty one is created.", kind, name))
	}

	return diags
}

func validateClickHouseDisk(diskPath path.Path, prior, planned *ClickHouseDiskModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior == nil || planned == nil {
		return diags
	}

	diags.Append(immutableString(diskPath.AtName("storage_class"), "storage_class", prior.StorageClass, planned.StorageClass)...)
	diags.Append(growOnlyInt64(diskPath.AtName("size"), "size", prior.Size, planned.Size)...)
	return diags
}

func immutableString(attrPath path.Path, name string, prior, planned types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	priorValue, priorOK := knownString(prior)
	plannedValue, plannedOK := knownString(planned)
	if !priorOK || !plannedOK {
		return diags
	}

	if priorValue != plannedValue {
		diags.AddAttributeError(attrPath, "Immutable Attribute",
			fmt.Sprintf("%s is immutable and cannot be modified after creation (%q -> %q).", name, priorValue, plannedValue))
	}

	return diags
}

// Zones are compared as a set: the API returns them in its own order and the
// attribute is immutable anyway, so a pure reorder is not a change.
func immutableList(attrPath path.Path, name string, prior, planned types.List) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior.IsNull() || prior.IsUnknown() || planned.IsNull() || planned.IsUnknown() {
		return diags
	}

	left, leftSettled := sortedStrings(prior)
	right, rightSettled := sortedStrings(planned)
	if !leftSettled || !rightSettled {
		return diags
	}

	if !sameStrings(left, right) {
		diags.AddAttributeError(attrPath, "Immutable Attribute", fmt.Sprintf("%s is immutable and cannot be modified after creation.", name))
	}

	return diags
}

func growOnlyInt64(attrPath path.Path, name string, prior, planned types.Int64) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior.IsNull() || prior.IsUnknown() || planned.IsNull() || planned.IsUnknown() {
		return diags
	}

	if planned.ValueInt64() < prior.ValueInt64() {
		diags.AddAttributeError(attrPath, "Volume Cannot Shrink",
			fmt.Sprintf("%s can only be increased, never decreased (%d -> %d).", name, prior.ValueInt64(), planned.ValueInt64()))
	}

	return diags
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

// settled is false when any element is unknown, which would otherwise shorten the
// slice and read as a removed zone.
func sortedStrings(list types.List) (values []string, settled bool) {
	values = make([]string, 0, len(list.Elements()))

	for _, element := range list.Elements() {
		s, ok := element.(types.String)
		if !ok || s.IsUnknown() || s.IsNull() {
			return nil, false
		}
		values = append(values, s.ValueString())
	}

	sort.Strings(values)
	return values, true
}

func knownString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", false
	}
	return value.ValueString(), true
}

// tfsdk.Config, tfsdk.Plan and tfsdk.State all satisfy this, so the same narrow
// read serves ValidateConfig and ModifyPlan.
type ClickHouseAttributeSource interface {
	GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics
}

// Reads only the two ClickHouse attributes: a full Get panics when any nested
// struct-pointer attribute elsewhere in the schema is unknown. Reports false when
// the values are not settled enough to validate.
func ReadClickHouse(ctx context.Context, src ClickHouseAttributeSource) ([]ClickHouseClusterModel, []ClickHouseKeeperModel, bool, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	var clusters []ClickHouseClusterModel
	ok, diags := ReadNestedList(ctx, src, clickHouseClustersPath, &clusters)
	allDiags.Append(diags...)
	if !ok {
		return nil, nil, false, allDiags
	}

	var keepers []ClickHouseKeeperModel
	ok, diags = ReadNestedList(ctx, src, clickHouseKeepersPath, &keepers)
	allDiags.Append(diags...)
	if !ok {
		return nil, nil, false, allDiags
	}

	return clusters, keepers, true, allDiags
}

// List-shaped node groups only; env_hcloud models them as a Set and needs its own read.
func ReadNodeGroups(ctx context.Context, src ClickHouseAttributeSource) ([]NodeGroupsModel, bool, diag.Diagnostics) {
	var nodeGroups []NodeGroupsModel
	ok, diags := ReadNestedList(ctx, src, nodeGroupsPath, &nodeGroups)
	return nodeGroups, ok, diags
}

func ReadNestedList[T any](ctx context.Context, src ClickHouseAttributeSource, attrPath path.Path, out *[]T) (bool, diag.Diagnostics) {
	var raw types.List
	diags := src.GetAttribute(ctx, attrPath, &raw)
	if diags.HasError() || raw.IsUnknown() {
		return false, diags
	}
	if raw.IsNull() {
		return true, diags
	}

	for _, element := range raw.Elements() {
		if element.IsUnknown() {
			return false, diags
		}
	}

	diags.Append(src.GetAttribute(ctx, attrPath, out)...)
	return !diags.HasError(), diags
}
