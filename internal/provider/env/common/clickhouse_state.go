package env

import (
	"slices"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Adopting clusters the config never declared would make the next plan delete them.
func ClickHouseClustersToModel(prior []ClickHouseClusterModel, specs []ClickHouseClusterSpec) ([]ClickHouseClusterModel, diag.Diagnostics) {
	if len(prior) == 0 {
		return nil, nil
	}
	return clickHouseClustersToModel(prior, specs)
}

// Data sources are read-only, so there is no configuration intent to protect.
func DataSourceClickHouseClustersToModel(specs []ClickHouseClusterSpec) ([]ClickHouseClusterModel, diag.Diagnostics) {
	return clickHouseClustersToModel(nil, specs)
}

// prior carries config/plan values so ordering and passwords survive a refresh.
func clickHouseClustersToModel(prior []ClickHouseClusterModel, specs []ClickHouseClusterSpec) ([]ClickHouseClusterModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	if len(specs) == 0 {
		return nil, allDiags
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseClusterModel) string { return m.Name.ValueString() },
		func(s ClickHouseClusterSpec) string { return s.Name },
	)

	priorByName := make(map[string]ClickHouseClusterModel, len(prior))
	for _, c := range prior {
		priorByName[c.Name.ValueString()] = c
	}

	clusters := make([]ClickHouseClusterModel, 0, len(specs))
	for _, s := range specs {
		p := priorByName[s.Name]

		zones, diags := ListToModel(reorderZones(p.Zones, s.Zones))
		allDiags.Append(diags...)

		users, diags := clickHouseUsersToModel(p.Users, s.Users)
		allDiags.Append(diags...)

		clusters = append(clusters, ClickHouseClusterModel{
			Name:            types.StringValue(s.Name),
			Mode:            types.StringValue(s.Mode),
			Image:           types.StringValue(s.Image),
			InstanceType:    types.StringValue(s.InstanceType),
			Zones:           zones,
			Shards:          types.Int64Value(s.Shards),
			Replicas:        types.Int64Value(s.Replicas),
			Stopped:         types.BoolValue(s.Stopped),
			Disk:            clickHouseDiskToModel(s.Disk),
			AdditionalDisks: clickHouseAdditionalDisksToModel(p.AdditionalDisks, s.AdditionalDisks),
			Keeper:          clickHouseKeeperRefToModel(p.Keeper, s.KeeperName),
			Settings:        clickHouseSettingsToModel(p.Settings, s.Settings),
			Profiles:        clickHouseProfilesToModel(p.Profiles, s.Profiles),
			Users:           users,
		})
	}

	return clusters, allDiags
}

func ClickHouseKeepersToModel(prior []ClickHouseKeeperModel, specs []ClickHouseKeeperSpec) ([]ClickHouseKeeperModel, diag.Diagnostics) {
	if len(prior) == 0 {
		return nil, nil
	}
	return clickHouseKeepersToModel(prior, specs)
}

func DataSourceClickHouseKeepersToModel(specs []ClickHouseKeeperSpec) ([]ClickHouseKeeperModel, diag.Diagnostics) {
	return clickHouseKeepersToModel(nil, specs)
}

func clickHouseKeepersToModel(prior []ClickHouseKeeperModel, specs []ClickHouseKeeperSpec) ([]ClickHouseKeeperModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	if len(specs) == 0 {
		return nil, allDiags
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseKeeperModel) string { return m.Name.ValueString() },
		func(s ClickHouseKeeperSpec) string { return s.Name },
	)

	priorByName := make(map[string]ClickHouseKeeperModel, len(prior))
	for _, k := range prior {
		priorByName[k.Name.ValueString()] = k
	}

	keepers := make([]ClickHouseKeeperModel, 0, len(specs))
	for _, s := range specs {
		zones, diags := ListToModel(reorderZones(priorByName[s.Name].Zones, s.Zones))
		allDiags.Append(diags...)

		keepers = append(keepers, ClickHouseKeeperModel{
			Name:         types.StringValue(s.Name),
			InstanceType: types.StringValue(s.InstanceType),
			Zones:        zones,
			HA:           types.BoolValue(s.HA),
			Stopped:      types.BoolValue(s.Stopped),
			Disk:         clickHouseDiskToModel(s.Disk),
		})
	}

	return keepers, allDiags
}

func clickHouseDiskToModel(spec ClickHouseDiskSpec) *ClickHouseDiskModel {
	return &ClickHouseDiskModel{
		Size:         types.Int64Value(spec.Size),
		StorageClass: types.StringValue(spec.StorageClass),
		Iops:         types.Int64Value(spec.Iops),
		Throughput:   types.Int64Value(spec.Throughput),
	}
}

func clickHouseAdditionalDisksToModel(prior []ClickHouseAdditionalDiskModel, specs []ClickHouseDiskSpec) []ClickHouseAdditionalDiskModel {
	if len(specs) == 0 {
		return nil
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseAdditionalDiskModel) string { return m.Name.ValueString() },
		func(s ClickHouseDiskSpec) string { return s.Name },
	)

	disks := make([]ClickHouseAdditionalDiskModel, 0, len(specs))
	for _, s := range specs {
		disks = append(disks, ClickHouseAdditionalDiskModel{
			Name:         types.StringValue(s.Name),
			Size:         types.Int64Value(s.Size),
			StorageClass: types.StringValue(s.StorageClass),
			Iops:         types.Int64Value(s.Iops),
			Throughput:   types.Int64Value(s.Throughput),
		})
	}

	return disks
}

// A null ref means detached; the name stays as configured since the API ignores it.
func clickHouseKeeperRefToModel(prior *ClickHouseKeeperRefModel, name *string) *ClickHouseKeeperRefModel {
	if name == nil {
		priorName := types.StringValue("")
		if prior != nil {
			priorName = prior.Name
		}
		return &ClickHouseKeeperRefModel{Enabled: types.BoolValue(false), Name: priorName}
	}

	return &ClickHouseKeeperRefModel{Enabled: types.BoolValue(true), Name: types.StringValue(*name)}
}

func clickHouseSettingsToModel(prior []ClickHouseSettingModel, specs []ClickHouseSettingSpec) []ClickHouseSettingModel {
	if len(specs) == 0 {
		return nil
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseSettingModel) string { return m.Key.ValueString() },
		func(s ClickHouseSettingSpec) string { return s.Key },
	)

	settings := make([]ClickHouseSettingModel, 0, len(specs))
	for _, s := range specs {
		// The API returns an empty value for a setting read from a secret.
		value := types.StringNull()
		if s.Value != "" {
			value = types.StringValue(s.Value)
		}

		settings = append(settings, ClickHouseSettingModel{
			Key:             types.StringValue(s.Key),
			Value:           value,
			ValueFromSecret: clickHouseSecretRefToModel(s.ValueFromSecret),
		})
	}

	return settings
}

func clickHouseProfilesToModel(prior []ClickHouseProfileModel, specs []ClickHouseProfileSpec) []ClickHouseProfileModel {
	if len(specs) == 0 {
		return nil
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseProfileModel) string { return m.Name.ValueString() },
		func(s ClickHouseProfileSpec) string { return s.Name },
	)

	priorByName := make(map[string][]ClickHouseSettingModel, len(prior))
	for _, p := range prior {
		priorByName[p.Name.ValueString()] = p.Settings
	}

	profiles := make([]ClickHouseProfileModel, 0, len(specs))
	for _, s := range specs {
		profiles = append(profiles, ClickHouseProfileModel{
			Name:     types.StringValue(s.Name),
			Settings: clickHouseSettingsToModel(priorByName[s.Name], s.Settings),
		})
	}

	return profiles
}

func clickHouseUsersToModel(prior []ClickHouseUserModel, specs []ClickHouseUserSpec) ([]ClickHouseUserModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	if len(specs) == 0 {
		return nil, allDiags
	}

	specs = ReorderByKey(prior, specs,
		func(m ClickHouseUserModel) string { return m.Name.ValueString() },
		func(s ClickHouseUserSpec) string { return s.Name },
	)

	priorByName := make(map[string]ClickHouseUserModel, len(prior))
	for _, u := range prior {
		priorByName[u.Name.ValueString()] = u
	}

	users := make([]ClickHouseUserModel, 0, len(specs))
	for _, s := range specs {
		// The config can never declare these, so state would not match the plan.
		if slices.Contains(clientsupport.ReservedClickHouseUsers, s.Name) {
			continue
		}

		p, hasPrior := priorByName[s.Name]

		allowedCIDRs, diags := nullableListToModel(s.AllowedCIDRs)
		allDiags.Append(diags...)
		databases, diags := nullableListToModel(s.Databases)
		allDiags.Append(diags...)

		// Passwords are never returned, so the config value is the only source of truth.
		passwordType := types.StringPointerValue(s.PasswordType)
		passwordValue := types.StringNull()
		if hasPrior {
			passwordType = p.PasswordType
			passwordValue = p.PasswordValue
		}

		users = append(users, ClickHouseUserModel{
			Name:                        types.StringValue(s.Name),
			Profile:                     types.StringValue(s.Profile),
			Quota:                       types.StringValue(s.Quota),
			AllowedCIDRs:                allowedCIDRs,
			Databases:                   databases,
			AccessManagement:            types.BoolValue(s.AccessManagement),
			NamedCollectionControl:      types.BoolValue(s.NamedCollectionControl),
			ShowNamedCollections:        types.BoolValue(s.ShowNamedCollections),
			ShowNamedCollectionsSecrets: types.BoolValue(s.ShowNamedCollectionsSecrets),
			PasswordType:                passwordType,
			PasswordValue:               passwordValue,
			PasswordValueFromSecret:     clickHouseSecretRefToModel(s.PasswordValueFromSecret),
		})
	}

	return users, allDiags
}

func clickHouseSecretRefToModel(spec *ClickHouseSecretRefSpec) *ClickHouseSecretRefModel {
	if spec == nil {
		return nil
	}

	return &ClickHouseSecretRefModel{
		Name: types.StringValue(spec.Name),
		Key:  types.StringValue(spec.Key),
	}
}

// The API returns zones in its own order; ctx-free because toModel has none.
func reorderZones(prior types.List, zones []string) []string {
	if prior.IsNull() || prior.IsUnknown() || len(zones) == 0 {
		return zones
	}

	ordered := make([]string, 0, len(zones))
	used := make(map[string]bool, len(zones))

	for _, element := range prior.Elements() {
		s, ok := element.(types.String)
		if !ok || s.IsUnknown() || s.IsNull() {
			continue
		}
		for _, zone := range zones {
			if !used[zone] && zone == s.ValueString() {
				ordered = append(ordered, zone)
				used[zone] = true
				break
			}
		}
	}

	for _, zone := range zones {
		if !used[zone] {
			ordered = append(ordered, zone)
		}
	}

	return ordered
}

// null vs [] fails the post-apply plan check on a plain Optional attribute.
func nullableListToModel(input []string) (types.List, diag.Diagnostics) {
	if len(input) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return ListToModel(input)
}
