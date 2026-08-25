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

// Cross-attribute rules the schema validators cannot express on their own.
func ValidateClickHouseConfig(clusters []ClickHouseClusterModel, keepers []ClickHouseKeeperModel) diag.Diagnostics {
	var diags diag.Diagnostics

	keeperNames := make(map[string]bool, len(keepers))
	for i, k := range keepers {
		name := k.Name.ValueString()
		if keeperNames[name] {
			diags.AddAttributeError(clickHouseKeepersPath.AtListIndex(i).AtName("name"), "Duplicate Keeper", fmt.Sprintf("Keeper %q is declared more than once.", name))
		}
		keeperNames[name] = true
	}

	clusterNames := make(map[string]bool, len(clusters))
	for i, c := range clusters {
		clusterPath := clickHouseClustersPath.AtListIndex(i)
		name := c.Name.ValueString()
		if clusterNames[name] {
			diags.AddAttributeError(clusterPath.AtName("name"), "Duplicate Cluster", fmt.Sprintf("ClickHouse cluster %q is declared more than once.", name))
		}
		clusterNames[name] = true

		diags.Append(validateClickHouseKeeperRef(clusterPath, c, keeperNames)...)
		diags.Append(validateUniqueNames(clusterPath.AtName("additional_disks"), "volume", c.AdditionalDisks,
			func(d ClickHouseAdditionalDiskModel) string { return d.Name.ValueString() })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("settings"), "setting", c.Settings,
			func(s ClickHouseSettingModel) string { return s.Key.ValueString() })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("profiles"), "profile", c.Profiles,
			func(p ClickHouseProfileModel) string { return p.Name.ValueString() })...)
		diags.Append(validateUniqueNames(clusterPath.AtName("users"), "user", c.Users,
			func(u ClickHouseUserModel) string { return u.Name.ValueString() })...)

		for j, p := range c.Profiles {
			diags.Append(validateUniqueNames(clusterPath.AtName("profiles").AtListIndex(j).AtName("settings"), "setting", p.Settings,
				func(s ClickHouseSettingModel) string { return s.Key.ValueString() })...)
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
	if !cluster.Keeper.Enabled.IsUnknown() && !cluster.Keeper.Enabled.ValueBool() {
		if cluster.Mode.ValueString() != string(sdk.ClickHouseClusterModeSpecSwarm) {
			diags.AddAttributeError(keeperPath.AtName("enabled"), "Keeper Required",
				fmt.Sprintf("Cluster %q must coordinate through a Keeper. Only a SWARM cluster may set `enabled` to `false`.", cluster.Name.ValueString()))
		}
		return diags
	}

	name := cluster.Keeper.Name
	if name.IsUnknown() || name.IsNull() {
		return diags
	}

	if !keeperNames[name.ValueString()] {
		diags.AddAttributeError(keeperPath.AtName("name"), "Unknown Keeper",
			fmt.Sprintf("Cluster %q references Keeper %q, which is not declared in `clickhouse_keepers`.", cluster.Name.ValueString(), name.ValueString()))
	}

	return diags
}

func validateUniqueNames[T any](attrPath path.Path, kind string, items []T, key func(T) string) diag.Diagnostics {
	var diags diag.Diagnostics
	seen := make(map[string]bool, len(items))

	for i, item := range items {
		k := key(item)
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

	for i, c := range planClusters {
		prior, ok := priorClusters[c.Name.ValueString()]
		if !ok {
			continue
		}

		clusterPath := clickHouseClustersPath.AtListIndex(i)
		diags.Append(immutableString(clusterPath.AtName("mode"), "mode", prior.Mode, c.Mode)...)
		diags.Append(immutableList(clusterPath.AtName("zones"), "zones", prior.Zones, c.Zones)...)
		diags.Append(validateClickHouseDisk(clusterPath.AtName("disk"), prior.Disk, c.Disk)...)

		priorDisks := make(map[string]ClickHouseAdditionalDiskModel, len(prior.AdditionalDisks))
		for _, d := range prior.AdditionalDisks {
			priorDisks[d.Name.ValueString()] = d
		}
		for j, d := range c.AdditionalDisks {
			priorDisk, ok := priorDisks[d.Name.ValueString()]
			if !ok {
				continue
			}
			diskPath := clusterPath.AtName("additional_disks").AtListIndex(j)
			diags.Append(immutableString(diskPath.AtName("storage_class"), "storage_class", priorDisk.StorageClass, d.StorageClass)...)
			diags.Append(growOnlyInt64(diskPath.AtName("size"), "size", priorDisk.Size, d.Size)...)
		}
	}

	priorKeepers := make(map[string]ClickHouseKeeperModel, len(stateKeepers))
	for _, k := range stateKeepers {
		priorKeepers[k.Name.ValueString()] = k
	}

	for i, k := range planKeepers {
		prior, ok := priorKeepers[k.Name.ValueString()]
		if !ok {
			continue
		}

		keeperPath := clickHouseKeepersPath.AtListIndex(i)
		diags.Append(immutableList(keeperPath.AtName("zones"), "zones", prior.Zones, k.Zones)...)
		diags.Append(validateClickHouseDisk(keeperPath.AtName("disk"), prior.Disk, k.Disk)...)

		if prior.HA.ValueBool() && !k.HA.IsUnknown() && !k.HA.ValueBool() {
			diags.AddAttributeError(keeperPath.AtName("ha"), "Immutable Attribute",
				fmt.Sprintf("Keeper %q is running as a highly-available ensemble and cannot be shrunk back to a single node.", k.Name.ValueString()))
		}
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
	if prior.IsNull() || prior.IsUnknown() || planned.IsNull() || planned.IsUnknown() {
		return diags
	}

	if prior.ValueString() != planned.ValueString() {
		diags.AddAttributeError(attrPath, "Immutable Attribute",
			fmt.Sprintf("%s is immutable and cannot be modified after creation (%q -> %q).", name, prior.ValueString(), planned.ValueString()))
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

	if !sameStringSet(prior, planned) {
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

func sameStringSet(a, b types.List) bool {
	left := sortedStrings(a)
	right := sortedStrings(b)
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

func sortedStrings(list types.List) []string {
	values := make([]string, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		s, ok := e.(types.String)
		if !ok || s.IsUnknown() || s.IsNull() {
			continue
		}
		values = append(values, s.ValueString())
	}

	sort.Strings(values)
	return values
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
	ok, diags := readClickHouseList(ctx, src, clickHouseClustersPath, &clusters)
	allDiags.Append(diags...)
	if !ok {
		return nil, nil, false, allDiags
	}

	var keepers []ClickHouseKeeperModel
	ok, diags = readClickHouseList(ctx, src, clickHouseKeepersPath, &keepers)
	allDiags.Append(diags...)
	if !ok {
		return nil, nil, false, allDiags
	}

	return clusters, keepers, true, allDiags
}

func readClickHouseList[T any](ctx context.Context, src ClickHouseAttributeSource, attrPath path.Path, out *[]T) (bool, diag.Diagnostics) {
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
