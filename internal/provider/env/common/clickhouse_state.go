package env

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	clientsupport "github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Only entries prior state already managed: adopting a console cluster would propose deleting it.
func ClickHouseClustersToModel(prior map[string]ClickHouseClusterModel, specs []ClickHouseClusterSpec) map[string]ClickHouseClusterModel {
	if prior == nil {
		return nil
	}

	var models map[string]ClickHouseClusterModel
	for _, s := range specs {
		if p, ok := prior[s.Name]; ok {
			if models == nil {
				models = make(map[string]ClickHouseClusterModel, len(prior))
			}
			models[s.Name] = clickHouseClusterToModel(p, s)
		}
	}

	return models
}

func ClickHouseKeepersToModel(prior map[string]ClickHouseKeeperModel, specs []ClickHouseKeeperSpec) map[string]ClickHouseKeeperModel {
	if prior == nil {
		return nil
	}

	var models map[string]ClickHouseKeeperModel
	for _, s := range specs {
		if p, ok := prior[s.Name]; ok {
			if models == nil {
				models = make(map[string]ClickHouseKeeperModel, len(prior))
			}
			models[s.Name] = clickHouseKeeperToModel(p, s)
		}
	}

	return models
}

// Import is the one read that takes ownership of every cluster and Keeper the environment has.
func AdoptClickHouseClusters(specs []ClickHouseClusterSpec) map[string]ClickHouseClusterModel {
	if len(specs) == 0 {
		return nil
	}

	models := make(map[string]ClickHouseClusterModel, len(specs))
	for _, s := range specs {
		models[s.Name] = clickHouseClusterToModel(ClickHouseClusterModel{}, s)
	}

	return models
}

func AdoptClickHouseKeepers(specs []ClickHouseKeeperSpec) map[string]ClickHouseKeeperModel {
	if len(specs) == 0 {
		return nil
	}

	models := make(map[string]ClickHouseKeeperModel, len(specs))
	for _, s := range specs {
		models[s.Name] = clickHouseKeeperToModel(ClickHouseKeeperModel{}, s)
	}

	return models
}

// A data source reports the environment as it is, reserved users included.
func ClickHouseClustersToDataSourceModel(specs []ClickHouseClusterSpec) map[string]ClickHouseClusterModel {
	models := AdoptClickHouseClusters(specs)
	for _, s := range specs {
		m := models[s.Name]
		m.Users = allUsersToModel(s.Users)
		models[s.Name] = m
	}

	return models
}

func ClickHouseKeepersToDataSourceModel(specs []ClickHouseKeeperSpec) map[string]ClickHouseKeeperModel {
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

func emptyMapLike[M any](prior map[string]M) map[string]M {
	if prior == nil {
		return nil
	}
	return map[string]M{}
}

func extraDisksToModel(prior map[string]ClickHouseExtraDiskModel, specs []ClickHouseNamedDiskSpec) map[string]ClickHouseExtraDiskModel {
	if len(specs) == 0 {
		return emptyMapLike(prior)
	}

	models := make(map[string]ClickHouseExtraDiskModel, len(specs))
	for _, s := range specs {
		models[s.Name] = ClickHouseExtraDiskModel{
			SizeGiB:        types.Int64Value(s.SizeGiB),
			Iops:           types.Int64Value(s.Iops),
			ThroughputMiBs: types.Int64Value(s.ThroughputMiBs),
		}
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

func profilesToModel(prior map[string]ClickHouseProfileModel, specs []ClickHouseProfileSpec) map[string]ClickHouseProfileModel {
	if len(specs) == 0 {
		return emptyMapLike(prior)
	}

	models := make(map[string]ClickHouseProfileModel, len(specs))
	for _, s := range specs {
		models[s.Name] = ClickHouseProfileModel{
			Settings: settingsToModel(prior[s.Name].Settings, s.Settings),
		}
	}

	return models
}

func quotasToModel(prior map[string]ClickHouseQuotaModel, specs []ClickHouseQuotaSpec) map[string]ClickHouseQuotaModel {
	if len(specs) == 0 {
		return emptyMapLike(prior)
	}

	models := make(map[string]ClickHouseQuotaModel, len(specs))
	for _, s := range specs {
		orderedIntervals := ReorderByKey(prior[s.Name].Intervals, s.Intervals,
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

		models[s.Name] = ClickHouseQuotaModel{
			KeyedBy:        types.StringValue(s.KeyedBy),
			IPv4PrefixBits: types.Int64PointerValue(s.IPv4PrefixBits),
			IPv6PrefixBits: types.Int64PointerValue(s.IPv6PrefixBits),
			Intervals:      intervals,
		}
	}

	return models
}

// Platform-injected users can never be declared, so keeping them would diverge state from plan.
func usersToModel(prior map[string]ClickHouseUserModel, specs []ClickHouseUserSpec) map[string]ClickHouseUserModel {
	var models map[string]ClickHouseUserModel
	for _, s := range specs {
		if slices.Contains(clientsupport.ReservedClickHouseUsers, s.Name) {
			continue
		}
		if models == nil {
			models = make(map[string]ClickHouseUserModel, len(specs))
		}
		models[s.Name] = userToModel(prior[s.Name], s)
	}

	if models == nil {
		return emptyMapLike(prior)
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

func allUsersToModel(specs []ClickHouseUserSpec) map[string]ClickHouseUserModel {
	if len(specs) == 0 {
		return nil
	}

	models := make(map[string]ClickHouseUserModel, len(specs))
	for _, s := range specs {
		models[s.Name] = userToModel(ClickHouseUserModel{}, s)
	}

	return models
}

// The API takes UInt64 quota limits, but Terraform's Int64 tops out at MaxInt64; a larger one would wrap negative.
func ValidateClickHouseQuotaLimits(specs []ClickHouseClusterSpec) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, c := range specs {
		for _, q := range c.Quotas {
			for _, i := range q.Intervals {
				limits := []struct {
					name  string
					value *uint64
				}{
					{"queries", i.Queries},
					{"query_selects", i.QuerySelects},
					{"query_inserts", i.QueryInserts},
					{"errors", i.Errors},
					{"result_rows", i.ResultRows},
					{"result_bytes", i.ResultBytes},
					{"read_rows", i.ReadRows},
					{"read_bytes", i.ReadBytes},
					{"written_bytes", i.WrittenBytes},
					{"failed_sequential_authentications", i.FailedSequentialAuthentications},
					{"queries_per_normalized_hash", i.QueriesPerNormalizedHash},
				}
				for _, l := range limits {
					if l.value != nil && *l.value > math.MaxInt64 {
						diags.AddError("ClickHouse Quota Limit Out of Range",
							fmt.Sprintf("Quota %q in cluster %q sets %s to %d, above the largest value Terraform can hold (%d). Lower it, or remove it for no limit, outside Terraform.", q.Name, c.Name, l.name, *l.value, int64(math.MaxInt64)))
					}
				}
			}
		}
	}
	return diags
}
