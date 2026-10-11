package env

import (
	"slices"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform only tracks what prior state already managed: a cluster created
// outside the configuration must not be adopted, or the next plan proposes
// deleting it along with its data.
func ClickHouseClustersToModel(prior []ClickHouseClusterModel, specs []ClickHouseClusterSpec) []ClickHouseClusterModel {
	if prior == nil {
		return nil
	}

	byName := make(map[string]ClickHouseClusterSpec, len(specs))
	for _, s := range specs {
		byName[s.Name] = s
	}

	var models []ClickHouseClusterModel
	for _, p := range prior {
		spec, ok := byName[p.Name.ValueString()]
		if !ok {
			continue
		}
		models = append(models, clickHouseClusterToModel(p, spec))
	}

	return models
}

func ClickHouseKeepersToModel(prior []ClickHouseKeeperModel, specs []ClickHouseKeeperSpec) []ClickHouseKeeperModel {
	if prior == nil {
		return nil
	}

	byName := make(map[string]ClickHouseKeeperSpec, len(specs))
	for _, s := range specs {
		byName[s.Name] = s
	}

	var models []ClickHouseKeeperModel
	for _, p := range prior {
		spec, ok := byName[p.Name.ValueString()]
		if !ok {
			continue
		}

		models = append(models, ClickHouseKeeperModel{
			Name:         types.StringValue(spec.Name),
			InstanceType: types.StringValue(spec.InstanceType),
			Zones:        reorderStringList(p.Zones, spec.Zones),
			Replicas:     types.Int64Value(spec.Replicas),
			Stopped:      types.BoolValue(spec.Stopped),
			Disk:         diskToModel(spec.Disk),
		})
	}

	return models
}

func clickHouseClusterToModel(prior ClickHouseClusterModel, spec ClickHouseClusterSpec) ClickHouseClusterModel {
	return ClickHouseClusterModel{
		Name:           types.StringValue(spec.Name),
		Image:          types.StringValue(spec.Image),
		InstanceType:   types.StringValue(spec.InstanceType),
		Zones:          reorderStringList(prior.Zones, spec.Zones),
		Shards:         types.Int64Value(spec.Shards),
		Replicas:       types.Int64Value(spec.Replicas),
		Stopped:        types.BoolValue(spec.Stopped),
		Keeper:         types.StringValue(spec.Keeper),
		Disk:           diskToModel(spec.Disk),
		ExtraDisks:     extraDisksToModel(prior.ExtraDisks, spec.ExtraDisks),
		Users:          usersToModel(prior.Users, spec.Users),
		Profiles:       profilesToModel(prior.Profiles, spec.Profiles),
		Quotas:         quotasToModel(prior.Quotas, spec.Quotas),
		ServerSettings: settingsToModel(prior.ServerSettings, spec.ServerSettings),
	}
}

func diskToModel(spec ClickHouseDiskSpec) *ClickHouseDiskModel {
	return &ClickHouseDiskModel{
		SizeGiB:        types.Int64Value(spec.SizeGiB),
		Iops:           types.Int64Value(spec.Iops),
		ThroughputMiBs: types.Int64Value(spec.ThroughputMiBs),
	}
}

func extraDisksToModel(prior []ClickHouseNamedDiskModel, specs []ClickHouseNamedDiskSpec) []ClickHouseNamedDiskModel {
	if specs == nil {
		return nil
	}

	ordered := reorderByName(prior, specs,
		func(m ClickHouseNamedDiskModel) string { return m.Name.ValueString() },
		func(s ClickHouseNamedDiskSpec) string { return s.Name },
	)

	models := make([]ClickHouseNamedDiskModel, 0, len(ordered))
	for _, s := range ordered {
		models = append(models, ClickHouseNamedDiskModel{
			Name:           types.StringValue(s.Name),
			SizeGiB:        types.Int64Value(s.SizeGiB),
			Iops:           types.Int64Value(s.Iops),
			ThroughputMiBs: types.Int64Value(s.ThroughputMiBs),
		})
	}

	return models
}

func settingsToModel(prior []ClickHouseSettingModel, specs []ClickHouseSettingSpec) []ClickHouseSettingModel {
	if specs == nil {
		return nil
	}

	ordered := reorderByName(prior, specs,
		func(m ClickHouseSettingModel) string { return m.Key.ValueString() },
		func(s ClickHouseSettingSpec) string { return s.Key },
	)

	models := make([]ClickHouseSettingModel, 0, len(ordered))
	for _, s := range ordered {
		models = append(models, ClickHouseSettingModel{
			Key:   types.StringValue(s.Key),
			Value: types.StringValue(s.Value),
		})
	}

	return models
}

func profilesToModel(prior []ClickHouseProfileModel, specs []ClickHouseProfileSpec) []ClickHouseProfileModel {
	if specs == nil {
		return nil
	}

	ordered := reorderByName(prior, specs,
		func(m ClickHouseProfileModel) string { return m.Name.ValueString() },
		func(s ClickHouseProfileSpec) string { return s.Name },
	)

	priorByName := make(map[string][]ClickHouseSettingModel, len(prior))
	for _, p := range prior {
		priorByName[p.Name.ValueString()] = p.Settings
	}

	models := make([]ClickHouseProfileModel, 0, len(ordered))
	for _, s := range ordered {
		models = append(models, ClickHouseProfileModel{
			Name:     types.StringValue(s.Name),
			Settings: settingsToModel(priorByName[s.Name], s.Settings),
		})
	}

	return models
}

func quotasToModel(prior []ClickHouseQuotaModel, specs []ClickHouseQuotaSpec) []ClickHouseQuotaModel {
	if specs == nil {
		return nil
	}

	ordered := reorderByName(prior, specs,
		func(m ClickHouseQuotaModel) string { return m.Name.ValueString() },
		func(s ClickHouseQuotaSpec) string { return s.Name },
	)

	models := make([]ClickHouseQuotaModel, 0, len(ordered))
	for _, s := range ordered {
		intervals := make([]ClickHouseQuotaIntervalModel, 0, len(s.Intervals))
		for _, i := range s.Intervals {
			intervals = append(intervals, ClickHouseQuotaIntervalModel{
				DurationSeconds:                 types.Int64Value(int64(i.DurationSeconds)),
				Queries:                         uint64ToModel(i.Queries),
				QuerySelects:                    uint64ToModel(i.QuerySelects),
				QueryInserts:                    uint64ToModel(i.QueryInserts),
				Errors:                          uint64ToModel(i.Errors),
				ResultRows:                      uint64ToModel(i.ResultRows),
				ResultBytes:                     uint64ToModel(i.ResultBytes),
				ReadRows:                        uint64ToModel(i.ReadRows),
				ReadBytes:                       uint64ToModel(i.ReadBytes),
				WrittenBytes:                    uint64ToModel(i.WrittenBytes),
				ExecutionTimeSeconds:            types.Float64PointerValue(i.ExecutionTimeSeconds),
				FailedSequentialAuthentications: uint64ToModel(i.FailedSequentialAuthentications),
				QueriesPerNormalizedHash:        uint64ToModel(i.QueriesPerNormalizedHash),
			})
		}

		models = append(models, ClickHouseQuotaModel{
			Name:           types.StringValue(s.Name),
			KeyedBy:        types.StringValue(s.KeyedBy),
			IPv4PrefixBits: types.Int64PointerValue(s.IPv4PrefixBits),
			IPv6PrefixBits: types.Int64PointerValue(s.IPv6PrefixBits),
			Intervals:      intervals,
		})
	}

	return models
}

// Reserved users are platform-injected: the configuration can never declare
// them, so keeping them would make the applied state differ from the plan.
func usersToModel(prior []ClickHouseUserModel, specs []ClickHouseUserSpec) []ClickHouseUserModel {
	if specs == nil {
		return nil
	}

	kept := make([]ClickHouseUserSpec, 0, len(specs))
	for _, s := range specs {
		if !slices.Contains(clientsupport.ReservedClickHouseUsers, s.Name) {
			kept = append(kept, s)
		}
	}

	ordered := reorderByName(prior, kept,
		func(m ClickHouseUserModel) string { return m.Name.ValueString() },
		func(s ClickHouseUserSpec) string { return s.Name },
	)

	priorByName := make(map[string]ClickHouseUserModel, len(prior))
	for _, p := range prior {
		priorByName[p.Name.ValueString()] = p
	}

	models := make([]ClickHouseUserModel, 0, len(ordered))
	for _, s := range ordered {
		p := priorByName[s.Name]

		// The API never returns enc_value, so it can only come from the config.
		encValue := types.StringNull()
		if p.Password != nil {
			encValue = p.Password.EncValue
		}

		var networks *ClickHouseUserNetworksModel
		if p.Networks != nil || len(s.Ips) > 0 {
			priorIps := types.ListNull(types.StringType)
			if p.Networks != nil {
				priorIps = p.Networks.Ips
			}
			networks = &ClickHouseUserNetworksModel{Ips: reorderStringList(priorIps, s.Ips)}
		}

		models = append(models, ClickHouseUserModel{
			Name:                        types.StringValue(s.Name),
			Profile:                     types.StringValue(s.Profile),
			Quota:                       types.StringValue(s.Quota),
			Networks:                    networks,
			AccessManagement:            types.BoolValue(s.AccessManagement),
			NamedCollectionControl:      types.BoolValue(s.NamedCollectionControl),
			ShowNamedCollectionsSecrets: types.BoolValue(s.ShowNamedCollectionsSecrets),
			Databases:                   reorderStringList(p.Databases, s.Databases),
			Grants:                      reorderStringList(p.Grants, s.Grants),
			Password: &ClickHouseUserPasswordModel{
				Type:     types.StringValue(s.PasswordType),
				EncValue: encValue,
			},
		})
	}

	return models
}

func uint64ToModel(value *uint64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}

// A null list must stay null: turning it into an empty one fails the
// post-apply check that the state matches the plan.
func reorderStringList(prior types.List, values []string) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}

	ordered := values
	if !prior.IsNull() && !prior.IsUnknown() {
		var priorValues []string
		for _, e := range prior.Elements() {
			if s, ok := e.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
				priorValues = append(priorValues, s.ValueString())
			}
		}
		ordered = reorderStrings(priorValues, values)
	}

	elements := make([]attr.Value, 0, len(ordered))
	for _, v := range ordered {
		elements = append(elements, types.StringValue(v))
	}

	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return list
}

func reorderStrings(prior, values []string) []string {
	remaining := make(map[string]bool, len(values))
	for _, v := range values {
		remaining[v] = true
	}

	ordered := make([]string, 0, len(values))
	for _, p := range prior {
		if remaining[p] {
			ordered = append(ordered, p)
			delete(remaining, p)
		}
	}
	for _, v := range values {
		if remaining[v] {
			ordered = append(ordered, v)
			delete(remaining, v)
		}
	}

	return ordered
}

// Entries the configuration never named go last, in the order the API gave them.
func reorderByName[M any, S any](prior []M, specs []S, priorKey func(M) string, specKey func(S) string) []S {
	index := make(map[string]int, len(specs))
	for i, s := range specs {
		index[specKey(s)] = i
	}

	ordered := make([]S, 0, len(specs))
	taken := make([]bool, len(specs))
	for _, p := range prior {
		if i, ok := index[priorKey(p)]; ok && !taken[i] {
			ordered = append(ordered, specs[i])
			taken[i] = true
		}
	}
	for i, s := range specs {
		if !taken[i] {
			ordered = append(ordered, s)
		}
	}

	return ordered
}
