package env

import (
	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
)

// gqlgenc emits one fragment type per env, so each env lowers its own into the shared spec structs.

// One constraint for all three generated disk fragment types, which share these getters.
type clickHouseDiskFragment interface {
	GetName() string
	GetSize() int64
	GetStorageClass() string
	GetIops() int64
	GetThroughput() int64
}

type clickHouseSecretRefFragment interface {
	GetName() string
	GetKey() string
}

func clickHouseClustersToSpec(clusters []*sdk.AWSEnvSpecFragment_ClickHouseClusters) []common.ClickHouseClusterSpec {
	var specs []common.ClickHouseClusterSpec

	for _, c := range clusters {
		if c == nil {
			continue
		}

		var additionalDisks []common.ClickHouseDiskSpec
		for _, d := range c.AdditionalDisks {
			if d != nil {
				additionalDisks = append(additionalDisks, clickHouseDiskToSpec(d))
			}
		}

		var keeperName *string
		if c.Keeper != nil {
			keeperName = &c.Keeper.Name
		}

		var profiles []common.ClickHouseProfileSpec
		for _, p := range c.Profiles {
			if p == nil {
				continue
			}

			var settings []common.ClickHouseSettingSpec
			for _, s := range p.Settings {
				if s != nil {
					settings = append(settings, common.ClickHouseSettingSpec{
						Key:             s.Key,
						Value:           s.Value,
						ValueFromSecret: clickHouseSecretRefToSpec(s.ValueFromSecret),
					})
				}
			}

			profiles = append(profiles, common.ClickHouseProfileSpec{Name: p.Name, Settings: settings})
		}

		var settings []common.ClickHouseSettingSpec
		for _, s := range c.Settings {
			if s != nil {
				settings = append(settings, common.ClickHouseSettingSpec{
					Key:             s.Key,
					Value:           s.Value,
					ValueFromSecret: clickHouseSecretRefToSpec(s.ValueFromSecret),
				})
			}
		}

		var users []common.ClickHouseUserSpec
		for _, u := range c.Users {
			if u == nil {
				continue
			}

			users = append(users, common.ClickHouseUserSpec{
				Name:                        u.Name,
				Profile:                     u.Profile,
				Quota:                       u.Quota,
				AllowedCIDRs:                u.AllowedCIDRs,
				Databases:                   u.Databases,
				AccessManagement:            u.AccessManagement,
				NamedCollectionControl:      u.NamedCollectionControl,
				ShowNamedCollections:        u.ShowNamedCollections,
				ShowNamedCollectionsSecrets: u.ShowNamedCollectionsSecrets,
				PasswordType:                (*string)(u.PasswordType),
				PasswordValueFromSecret:     clickHouseSecretRefToSpec(u.PasswordValueFromSecret),
			})
		}

		specs = append(specs, common.ClickHouseClusterSpec{
			Name:            c.Name,
			Mode:            string(c.Mode),
			Image:           c.Image,
			InstanceType:    c.InstanceType,
			Zones:           c.Zones,
			Shards:          c.Shards,
			Replicas:        c.Replicas,
			Stopped:         c.Stopped,
			Disk:            clickHouseDiskToSpec(&c.Disk),
			AdditionalDisks: additionalDisks,
			KeeperName:      keeperName,
			Settings:        settings,
			Profiles:        profiles,
			Users:           users,
		})
	}

	return specs
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
			HA:           k.Ha,
			Stopped:      k.Stopped,
			Disk:         clickHouseDiskToSpec(&k.Disk),
		})
	}

	return specs
}

func clickHouseDiskToSpec(disk clickHouseDiskFragment) common.ClickHouseDiskSpec {
	return common.ClickHouseDiskSpec{
		Name:         disk.GetName(),
		Size:         disk.GetSize(),
		StorageClass: disk.GetStorageClass(),
		Iops:         disk.GetIops(),
		Throughput:   disk.GetThroughput(),
	}
}

// The nil check has to happen before the pointer is boxed into the interface.
func clickHouseSecretRefToSpec[T any, PT interface {
	*T
	clickHouseSecretRefFragment
}](ref PT) *common.ClickHouseSecretRefSpec {
	if ref == nil {
		return nil
	}

	return &common.ClickHouseSecretRefSpec{Name: ref.GetName(), Key: ref.GetKey()}
}
