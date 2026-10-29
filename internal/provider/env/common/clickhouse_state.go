package env

import (
	"slices"
	"strconv"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Only entries prior state already managed: adopting a console cluster would propose deleting it.
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
		models = append(models, clickHouseKeeperToModel(p, spec))
	}

	return models
}

// Import is the one read that takes ownership of every cluster and Keeper the environment has.
func AdoptClickHouseClusters(specs []ClickHouseClusterSpec) []ClickHouseClusterModel {
	if len(specs) == 0 {
		return nil
	}

	models := make([]ClickHouseClusterModel, 0, len(specs))
	for _, s := range specs {
		models = append(models, clickHouseClusterToModel(ClickHouseClusterModel{}, s))
	}

	return models
}

func AdoptClickHouseKeepers(specs []ClickHouseKeeperSpec) []ClickHouseKeeperModel {
	if len(specs) == 0 {
		return nil
	}

	models := make([]ClickHouseKeeperModel, 0, len(specs))
	for _, s := range specs {
		models = append(models, clickHouseKeeperToModel(ClickHouseKeeperModel{}, s))
	}

	return models
}

// A data source reports the environment as it is, reserved users included.
func ClickHouseClustersToDataSourceModel(specs []ClickHouseClusterSpec) []ClickHouseClusterModel {
	models := AdoptClickHouseClusters(specs)
	for i, s := range specs {
		models[i].Users = allUsersToModel(s.Users)
	}

	return models
}

func ClickHouseKeepersToDataSourceModel(specs []ClickHouseKeeperSpec) []ClickHouseKeeperModel {
	return AdoptClickHouseKeepers(specs)
}

// The API reports "no Keeper" as "", which the config expresses by omitting the attribute.
func keeperToModel(name string) types.String {
	if name == "" {
		return types.StringNull()
	}
	return types.StringValue(name)
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
		Keeper:         keeperToModel(spec.Keeper),
		Disk:           diskToModel(spec.Disk),
		ExtraDisks:     extraDisksToModel(prior.ExtraDisks, spec.ExtraDisks),
		Users:          usersToModel(prior.Users, spec.Users),
		Profiles:       profilesToModel(prior.Profiles, spec.Profiles),
		Quotas:         quotasToModel(prior.Quotas, spec.Quotas),
		ServerSettings: settingsToModel(prior.ServerSettings, spec.ServerSettings),
	}
}

func clickHouseKeeperToModel(prior ClickHouseKeeperModel, spec ClickHouseKeeperSpec) ClickHouseKeeperModel {
	return ClickHouseKeeperModel{
		Name:         types.StringValue(spec.Name),
		InstanceType: types.StringValue(spec.InstanceType),
		Zones:        reorderStringList(prior.Zones, spec.Zones),
		Replicas:     types.Int64Value(spec.Replicas),
		Stopped:      types.BoolValue(spec.Stopped),
		Disk:         diskToModel(spec.Disk),
	}
}

func diskToModel(spec ClickHouseDiskSpec) *ClickHouseDiskModel {
	return &ClickHouseDiskModel{
		SizeGiB:        types.Int64Value(spec.SizeGiB),
		Iops:           types.Int64Value(spec.Iops),
		ThroughputMiBs: types.Int64Value(spec.ThroughputMiBs),
	}
}

// A null plan must read back null and an empty one empty, or the post-apply check fails.
func emptyLike[M any](prior []M) []M {
	if prior == nil {
		return nil
	}
	return []M{}
}

func extraDisksToModel(prior []ClickHouseNamedDiskModel, specs []ClickHouseNamedDiskSpec) []ClickHouseNamedDiskModel {
	if len(specs) == 0 {
		return emptyLike(prior)
	}

	ordered := ReorderByKey(prior, specs,
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
	if len(specs) == 0 {
		return emptyLike(prior)
	}

	ordered := ReorderByKey(prior, specs,
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
	if len(specs) == 0 {
		return emptyLike(prior)
	}

	ordered := ReorderByKey(prior, specs,
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
	if len(specs) == 0 {
		return emptyLike(prior)
	}

	ordered := ReorderByKey(prior, specs,
		func(m ClickHouseQuotaModel) string { return m.Name.ValueString() },
		func(s ClickHouseQuotaSpec) string { return s.Name },
	)

	priorByName := make(map[string][]ClickHouseQuotaIntervalModel, len(prior))
	for _, p := range prior {
		priorByName[p.Name.ValueString()] = p.Intervals
	}

	models := make([]ClickHouseQuotaModel, 0, len(ordered))
	for _, s := range ordered {
		orderedIntervals := ReorderByKey(priorByName[s.Name], s.Intervals,
			func(m ClickHouseQuotaIntervalModel) string {
				return strconv.FormatInt(m.DurationSeconds.ValueInt64(), 10)
			},
			func(i ClickHouseQuotaIntervalSpec) string { return strconv.FormatUint(i.DurationSeconds, 10) },
		)

		intervals := make([]ClickHouseQuotaIntervalModel, 0, len(orderedIntervals))
		for _, i := range orderedIntervals {
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

// Platform-injected users can never be declared, so keeping them would diverge state from plan.
func usersToModel(prior []ClickHouseUserModel, specs []ClickHouseUserSpec) []ClickHouseUserModel {
	kept := make([]ClickHouseUserSpec, 0, len(specs))
	for _, s := range specs {
		if !slices.Contains(clientsupport.ReservedClickHouseUsers, s.Name) {
			kept = append(kept, s)
		}
	}

	if len(kept) == 0 {
		return emptyLike(prior)
	}

	ordered := ReorderByKey(prior, kept,
		func(m ClickHouseUserModel) string { return m.Name.ValueString() },
		func(s ClickHouseUserSpec) string { return s.Name },
	)

	priorByName := make(map[string]ClickHouseUserModel, len(prior))
	for _, p := range prior {
		priorByName[p.Name.ValueString()] = p
	}

	models := make([]ClickHouseUserModel, 0, len(ordered))
	for _, s := range ordered {
		models = append(models, userToModel(priorByName[s.Name], s))
	}

	return models
}

func userToModel(prior ClickHouseUserModel, spec ClickHouseUserSpec) ClickHouseUserModel {
	// The API never returns enc_value, so it can only come from the config.
	encValue := types.StringNull()
	if prior.Password != nil {
		encValue = prior.Password.EncValue
	}

	var networks *ClickHouseUserNetworksModel
	if prior.Networks != nil || len(spec.Ips) > 0 {
		priorIps := types.ListNull(types.StringType)
		if prior.Networks != nil {
			priorIps = prior.Networks.Ips
		}
		networks = &ClickHouseUserNetworksModel{Ips: reorderStringList(priorIps, spec.Ips)}
	}

	return ClickHouseUserModel{
		Name:                        types.StringValue(spec.Name),
		Profile:                     types.StringValue(spec.Profile),
		Quota:                       types.StringValue(spec.Quota),
		Networks:                    networks,
		AccessManagement:            types.BoolValue(spec.AccessManagement),
		NamedCollectionControl:      types.BoolValue(spec.NamedCollectionControl),
		ShowNamedCollectionsSecrets: types.BoolValue(spec.ShowNamedCollectionsSecrets),
		Databases:                   reorderStringList(prior.Databases, spec.Databases),
		Grants:                      reorderStringList(prior.Grants, spec.Grants),
		Password: &ClickHouseUserPasswordModel{
			Type:     types.StringValue(spec.PasswordType),
			EncValue: encValue,
		},
	}
}

func uint64ToModel(value *uint64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}

// Null and empty echoes mirror the prior value, or the post-apply plan/state check fails.
func reorderStringList(prior types.List, values []string) types.List {
	if len(values) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.ListNull(types.StringType)
		}
		return types.ListValueMust(types.StringType, []attr.Value{})
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

	return types.ListValueMust(types.StringType, elements)
}

func allUsersToModel(specs []ClickHouseUserSpec) []ClickHouseUserModel {
	if len(specs) == 0 {
		return nil
	}

	models := make([]ClickHouseUserModel, 0, len(specs))
	for _, s := range specs {
		models = append(models, userToModel(ClickHouseUserModel{}, s))
	}

	return models
}
