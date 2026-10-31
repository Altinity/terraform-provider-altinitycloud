package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// gqlgenc emits one fragment and input family per env; this adapts AWS's to the shared structs.

func clickHouseClustersToSpec(clusters []*sdk.AWSEnvSpecFragment_ClickHouseClusters) ([]common.ClickHouseClusterSpec, diag.Diagnostics) {
	var specs []common.ClickHouseClusterSpec

	for _, c := range clusters {
		if c == nil {
			continue
		}

		var extraDisks []common.ClickHouseNamedDiskSpec
		for _, d := range c.ExtraDisks {
			if d != nil {
				extraDisks = append(extraDisks, common.ClickHouseNamedDiskSpec{
					Name:           d.Name,
					SizeGiB:        d.SizeInGiB,
					Iops:           d.Iops,
					ThroughputMiBs: d.ThroughputInMiBps,
				})
			}
		}

		var users []common.ClickHouseUserSpec
		for _, u := range c.Users {
			if u == nil {
				continue
			}

			var databases []string
			for _, d := range u.Databases {
				if d != nil {
					databases = append(databases, d.Name)
				}
			}

			users = append(users, common.ClickHouseUserSpec{
				Name:                        u.Name,
				Profile:                     u.Profile,
				Quota:                       u.Quota,
				Ips:                         u.Networks.Ips,
				AccessManagement:            u.AccessManagement,
				NamedCollectionControl:      u.NamedCollectionControl,
				ShowNamedCollectionsSecrets: u.ShowNamedCollectionsSecrets,
				Databases:                   databases,
				Grants:                      u.Grants,
				PasswordType:                string(u.Password.Type),
			})
		}

		var profiles []common.ClickHouseProfileSpec
		for _, p := range c.Profiles {
			if p == nil {
				continue
			}

			var settings []common.ClickHouseSettingSpec
			for _, s := range p.Settings {
				if s != nil {
					settings = append(settings, common.ClickHouseSettingSpec{Key: s.Key, Value: s.Value})
				}
			}

			profiles = append(profiles, common.ClickHouseProfileSpec{Name: p.Name, Settings: settings})
		}

		var quotas []common.ClickHouseQuotaSpec
		for _, q := range c.Quotas {
			if q == nil {
				continue
			}

			var intervals []common.ClickHouseQuotaIntervalSpec
			for _, i := range q.Intervals {
				if i != nil {
					intervals = append(intervals, common.ClickHouseQuotaIntervalSpec{
						DurationSeconds:                 i.Duration,
						Queries:                         i.Queries,
						QuerySelects:                    i.QuerySelects,
						QueryInserts:                    i.QueryInserts,
						Errors:                          i.Errors,
						ResultRows:                      i.ResultRows,
						ResultBytes:                     i.ResultBytes,
						ReadRows:                        i.ReadRows,
						ReadBytes:                       i.ReadBytes,
						WrittenBytes:                    i.WrittenBytes,
						ExecutionTimeSeconds:            i.ExecutionTime,
						FailedSequentialAuthentications: i.FailedSequentialAuthentications,
						QueriesPerNormalizedHash:        i.QueriesPerNormalizedHash,
					})
				}
			}

			quotas = append(quotas, common.ClickHouseQuotaSpec{
				Name:           q.Name,
				KeyedBy:        string(q.KeyedBy),
				IPv4PrefixBits: q.Ipv4PrefixBits,
				IPv6PrefixBits: q.Ipv6PrefixBits,
				Intervals:      intervals,
			})
		}

		var serverSettings []common.ClickHouseSettingSpec
		for _, s := range c.ServerSettings {
			if s != nil {
				serverSettings = append(serverSettings, common.ClickHouseSettingSpec{Key: s.Key, Value: s.Value})
			}
		}

		specs = append(specs, common.ClickHouseClusterSpec{
			Name:         c.Name,
			Image:        c.Image,
			InstanceType: c.InstanceType,
			Zones:        c.Zones,
			Shards:       c.Shards,
			Replicas:     c.Replicas,
			Stopped:      c.Stopped,
			Keeper:       c.Keeper,
			Disk: common.ClickHouseDiskSpec{
				SizeGiB:        c.Disk.SizeInGiB,
				Iops:           c.Disk.Iops,
				ThroughputMiBs: c.Disk.ThroughputInMiBps,
			},
			ExtraDisks:     extraDisks,
			Users:          users,
			Profiles:       profiles,
			Quotas:         quotas,
			ServerSettings: serverSettings,
		})
	}

	return specs, common.ValidateClickHouseQuotaLimits(specs)
}

func clickHouseKeepersToSpec(keepers []*sdk.AWSEnvSpecFragment_ClickHouseKeepers) []common.ClickHouseKeeperSpec {
	var specs []common.ClickHouseKeeperSpec

	for _, k := range keepers {
		if k == nil {
			continue
		}

		specs = append(specs, common.ClickHouseKeeperSpec{
			Name:         k.Name,
			InstanceType: k.InstanceType,
			Zones:        k.Zones,
			Replicas:     k.Replicas,
			Stopped:      k.Stopped,
			Disk: common.ClickHouseDiskSpec{
				SizeGiB:        k.Disk.SizeInGiB,
				Iops:           k.Disk.Iops,
				ThroughputMiBs: k.Disk.ThroughputInMiBps,
			},
		})
	}

	return specs
}

func clickHouseClustersToSDK(ctx context.Context, clusters map[string]common.ClickHouseClusterModel) ([]*sdk.AWSEnvClickHouseClusterSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.AWSEnvClickHouseClusterSpecInput

	for _, name := range common.SortedKeys(clusters) {
		c := clusters[name]
		zones, diags := common.StringListToSDK(ctx, c.Zones)
		allDiags.Append(diags...)

		users, diags := clickHouseUsersToSDK(ctx, c.Users)
		allDiags.Append(diags...)

		var extraDisks []*sdk.AWSEnvClickHouseNamedDiskSpecInput
		for _, diskName := range common.SortedKeys(c.ExtraDisks) {
			d := c.ExtraDisks[diskName]
			extraDisks = append(extraDisks, &sdk.AWSEnvClickHouseNamedDiskSpecInput{
				Name:              diskName,
				SizeInGiB:         d.SizeGiB.ValueInt64(),
				Iops:              common.KnownInt64Pointer(d.Iops),
				ThroughputInMiBps: common.KnownInt64Pointer(d.ThroughputMiBs),
			})
		}

		var profiles []*sdk.AWSEnvClickHouseProfileSpecInput
		for _, profileName := range common.SortedKeys(c.Profiles) {
			p := c.Profiles[profileName]
			var settings []*sdk.AWSEnvClickHouseProfileSettingSpecInput
			for _, s := range p.Settings {
				settings = append(settings, &sdk.AWSEnvClickHouseProfileSettingSpecInput{
					Key:   s.Key.ValueString(),
					Value: common.KnownStringPointer(s.Value),
				})
			}
			profiles = append(profiles, &sdk.AWSEnvClickHouseProfileSpecInput{
				Name:     profileName,
				Settings: settings,
			})
		}

		var serverSettings []*sdk.AWSEnvClickHouseServerSettingSpecInput
		for _, s := range c.ServerSettings {
			serverSettings = append(serverSettings, &sdk.AWSEnvClickHouseServerSettingSpecInput{
				Key:   s.Key.ValueString(),
				Value: common.KnownStringPointer(s.Value),
			})
		}

		input = append(input, &sdk.AWSEnvClickHouseClusterSpecInput{
			Name:           name,
			Image:          common.KnownStringPointer(c.Image),
			InstanceType:   c.InstanceType.ValueString(),
			Zones:          zones,
			Shards:         common.KnownInt64Pointer(c.Shards),
			Replicas:       common.KnownInt64Pointer(c.Replicas),
			Stopped:        common.KnownBoolPointer(c.Stopped),
			Keeper:         common.KnownStringPointer(c.Keeper),
			Disk:           clickHouseDiskToSDK(c.Disk),
			ExtraDisks:     extraDisks,
			Users:          users,
			Profiles:       profiles,
			Quotas:         common.ClickHouseQuotasToSDK(c.Quotas),
			ServerSettings: serverSettings,
		})
	}

	return input, allDiags
}

func clickHouseKeepersToSDK(ctx context.Context, keepers map[string]common.ClickHouseKeeperModel) ([]*sdk.AWSEnvClickHouseKeeperSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.AWSEnvClickHouseKeeperSpecInput

	for _, name := range common.SortedKeys(keepers) {
		k := keepers[name]
		zones, diags := common.StringListToSDK(ctx, k.Zones)
		allDiags.Append(diags...)

		var disk *sdk.AWSEnvClickHouseKeeperDiskSpecInput
		if k.Disk != nil {
			disk = &sdk.AWSEnvClickHouseKeeperDiskSpecInput{
				SizeInGiB:         k.Disk.SizeGiB.ValueInt64(),
				Iops:              common.KnownInt64Pointer(k.Disk.Iops),
				ThroughputInMiBps: common.KnownInt64Pointer(k.Disk.ThroughputMiBs),
			}
		}

		input = append(input, &sdk.AWSEnvClickHouseKeeperSpecInput{
			Name:         name,
			InstanceType: k.InstanceType.ValueString(),
			Zones:        zones,
			Replicas:     common.KnownInt64Pointer(k.Replicas),
			Stopped:      common.KnownBoolPointer(k.Stopped),
			Disk:         disk,
		})
	}

	return input, allDiags
}

// A nil patch leaves the API's entries untouched; one is sent only when Terraform owns a change.
func clickHouseClustersPatch(replace []*sdk.AWSEnvClickHouseClusterSpecInput, remove []string) *sdk.AWSEnvClickHouseClustersPatchInput {
	if replace == nil && remove == nil {
		return nil
	}
	return &sdk.AWSEnvClickHouseClustersPatchInput{Replace: replace, Delete: remove}
}

func clickHouseKeepersPatch(replace []*sdk.AWSEnvClickHouseKeeperSpecInput, remove []string) *sdk.AWSEnvClickHouseKeepersPatchInput {
	if replace == nil && remove == nil {
		return nil
	}
	return &sdk.AWSEnvClickHouseKeepersPatchInput{Replace: replace, Delete: remove}
}

func clickHouseDiskToSDK(disk *common.ClickHouseDiskModel) *sdk.AWSEnvClickHouseClusterDiskSpecInput {
	if disk == nil {
		return nil
	}

	return &sdk.AWSEnvClickHouseClusterDiskSpecInput{
		SizeInGiB:         disk.SizeGiB.ValueInt64(),
		Iops:              common.KnownInt64Pointer(disk.Iops),
		ThroughputInMiBps: common.KnownInt64Pointer(disk.ThroughputMiBs),
	}
}

func clickHouseUsersToSDK(ctx context.Context, users map[string]common.ClickHouseUserModel) ([]*sdk.AWSEnvClickHouseUserSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.AWSEnvClickHouseUserSpecInput

	for _, name := range common.SortedKeys(users) {
		u := users[name]
		grants, diags := common.StringListToSDK(ctx, u.Grants)
		allDiags.Append(diags...)
		databaseNames, diags := common.StringListToSDK(ctx, u.Databases)
		allDiags.Append(diags...)

		var networks *sdk.AWSEnvClickHouseUserNetworksSpecInput
		if u.Networks != nil {
			ips, diags := common.StringListToSDK(ctx, u.Networks.Ips)
			allDiags.Append(diags...)
			networks = &sdk.AWSEnvClickHouseUserNetworksSpecInput{Ips: ips}
		}

		var databases []*sdk.AWSEnvClickHouseUserDatabaseSpecInput
		for _, name := range databaseNames {
			databases = append(databases, &sdk.AWSEnvClickHouseUserDatabaseSpecInput{Name: name})
		}

		var password *sdk.AWSEnvClickHouseUserPasswordSpecInput
		if u.Password != nil {
			password = &sdk.AWSEnvClickHouseUserPasswordSpecInput{
				Type:     sdk.AWSEnvClickHouseUserPasswordType(u.Password.Type.ValueString()),
				EncValue: common.KnownStringPointer(u.Password.EncValue),
			}
		}

		input = append(input, &sdk.AWSEnvClickHouseUserSpecInput{
			Name:                        name,
			Profile:                     common.KnownStringPointer(u.Profile),
			Quota:                       common.KnownStringPointer(u.Quota),
			Networks:                    networks,
			AccessManagement:            common.KnownBoolPointer(u.AccessManagement),
			NamedCollectionControl:      common.KnownBoolPointer(u.NamedCollectionControl),
			ShowNamedCollectionsSecrets: common.KnownBoolPointer(u.ShowNamedCollectionsSecrets),
			Databases:                   databases,
			Grants:                      grants,
			Password:                    password,
		})
	}

	return input, allDiags
}
