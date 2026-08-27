package env

import (
	"context"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API names a cluster's main volume and a Keeper's volume `default`.
const ClickHouseDefaultDiskName = "default"

func ClickHouseClustersToSDK(ctx context.Context, clusters []ClickHouseClusterModel) ([]*sdk.ClickHouseClusterCreateSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.ClickHouseClusterCreateSpecInput

	for _, c := range clusters {
		zones, diags := stringListToSDK(ctx, c.Zones)
		allDiags.Append(diags...)

		users, diags := clickHouseUsersToSDK(ctx, c.Users)
		allDiags.Append(diags...)

		var additionalDisks []*sdk.ClickHouseDiskCreateSpecInput
		for _, d := range c.AdditionalDisks {
			additionalDisks = append(additionalDisks, &sdk.ClickHouseDiskCreateSpecInput{
				Name:         d.Name.ValueString(),
				Size:         knownInt64Pointer(d.Size),
				StorageClass: knownStringPointer(d.StorageClass),
				Iops:         knownInt64Pointer(d.Iops),
				Throughput:   knownInt64Pointer(d.Throughput),
			})
		}

		input = append(input, &sdk.ClickHouseClusterCreateSpecInput{
			Name:            c.Name.ValueString(),
			Mode:            (*sdk.ClickHouseClusterModeSpec)(knownStringPointer(c.Mode)),
			Image:           c.Image.ValueString(),
			InstanceType:    c.InstanceType.ValueString(),
			Zones:           zones,
			Shards:          c.Shards.ValueInt64(),
			Replicas:        c.Replicas.ValueInt64(),
			Stopped:         knownBoolPointer(c.Stopped),
			Disk:            clickHouseDiskToSDK(ClickHouseDefaultDiskName, c.Disk),
			AdditionalDisks: additionalDisks,
			Keeper:          clickHouseKeeperRefToSDK(c.Keeper),
			Settings:        clickHouseSettingsToSDK(c.Settings),
			Profiles:        clickHouseProfilesToSDK(c.Profiles),
			Users:           users,
		})
	}

	return input, allDiags
}

// `mode` and `zones` are immutable, so the update input carries neither.
func ClickHouseClustersToUpdateSDK(ctx context.Context, clusters []ClickHouseClusterModel) ([]*sdk.ClickHouseClusterUpdateSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.ClickHouseClusterUpdateSpecInput

	for _, c := range clusters {
		users, diags := clickHouseUsersToSDK(ctx, c.Users)
		allDiags.Append(diags...)

		var additionalDisks []*sdk.ClickHouseDiskUpdateSpecInput
		for _, d := range c.AdditionalDisks {
			additionalDisks = append(additionalDisks, &sdk.ClickHouseDiskUpdateSpecInput{
				Name:       d.Name.ValueString(),
				Size:       knownInt64Pointer(d.Size),
				Iops:       knownInt64Pointer(d.Iops),
				Throughput: knownInt64Pointer(d.Throughput),
			})
		}

		input = append(input, &sdk.ClickHouseClusterUpdateSpecInput{
			Name:            c.Name.ValueString(),
			Image:           knownStringPointer(c.Image),
			InstanceType:    knownStringPointer(c.InstanceType),
			Shards:          knownInt64Pointer(c.Shards),
			Replicas:        knownInt64Pointer(c.Replicas),
			Stopped:         knownBoolPointer(c.Stopped),
			Disk:            clickHouseDiskToUpdateSDK(ClickHouseDefaultDiskName, c.Disk),
			AdditionalDisks: additionalDisks,
			Keeper:          clickHouseKeeperRefToSDK(c.Keeper),
			Settings:        clickHouseSettingsToSDK(c.Settings),
			Profiles:        clickHouseProfilesToUpdateSDK(c.Profiles),
			Users:           users,
		})
	}

	return input, allDiags
}

func ClickHouseKeepersToSDK(ctx context.Context, keepers []ClickHouseKeeperModel) ([]*sdk.ClickHouseKeeperCreateSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.ClickHouseKeeperCreateSpecInput

	for _, k := range keepers {
		zones, diags := stringListToSDK(ctx, k.Zones)
		allDiags.Append(diags...)

		input = append(input, &sdk.ClickHouseKeeperCreateSpecInput{
			Name:         k.Name.ValueString(),
			InstanceType: k.InstanceType.ValueString(),
			Zones:        zones,
			Ha:           knownBoolPointer(k.HA),
			Stopped:      knownBoolPointer(k.Stopped),
			Disk:         clickHouseDiskToSDK(ClickHouseDefaultDiskName, k.Disk),
		})
	}

	return input, allDiags
}

// `zones` is immutable, so the update input carries none.
func ClickHouseKeepersToUpdateSDK(keepers []ClickHouseKeeperModel) []*sdk.ClickHouseKeeperUpdateSpecInput {
	var input []*sdk.ClickHouseKeeperUpdateSpecInput

	for _, k := range keepers {
		input = append(input, &sdk.ClickHouseKeeperUpdateSpecInput{
			Name:         k.Name.ValueString(),
			InstanceType: knownStringPointer(k.InstanceType),
			Ha:           knownBoolPointer(k.HA),
			Stopped:      knownBoolPointer(k.Stopped),
			Disk:         clickHouseDiskToUpdateSDK(ClickHouseDefaultDiskName, k.Disk),
		})
	}

	return input
}

// An entry dropped from the config is kept by the API unless it is named here.
func ClickHouseClusterNamesToDelete(prior, planned []ClickHouseClusterModel) []string {
	return namesToDelete(prior, planned,
		func(m ClickHouseClusterModel) string { return m.Name.ValueString() },
		func(m ClickHouseClusterModel) string { return m.Name.ValueString() },
	)
}

func ClickHouseKeeperNamesToDelete(prior, planned []ClickHouseKeeperModel) []string {
	return namesToDelete(prior, planned,
		func(m ClickHouseKeeperModel) string { return m.Name.ValueString() },
		func(m ClickHouseKeeperModel) string { return m.Name.ValueString() },
	)
}

// Diffs prior state against the patch entries already built from the plan.
func ApplyClickHouseClusterNestedDeletes(updates []*sdk.ClickHouseClusterUpdateSpecInput, prior []ClickHouseClusterModel) {
	priorByName := make(map[string]ClickHouseClusterModel, len(prior))
	for _, c := range prior {
		priorByName[c.Name.ValueString()] = c
	}

	for _, u := range updates {
		p, ok := priorByName[u.Name]
		if !ok {
			continue
		}

		u.AdditionalDisksToDelete = namesToDelete(p.AdditionalDisks, u.AdditionalDisks,
			func(m ClickHouseAdditionalDiskModel) string { return m.Name.ValueString() },
			func(s *sdk.ClickHouseDiskUpdateSpecInput) string { return s.Name },
		)
		u.SettingsToDelete = namesToDelete(p.Settings, u.Settings,
			func(m ClickHouseSettingModel) string { return m.Key.ValueString() },
			func(s *sdk.ClickHouseSettingSpecInput) string { return s.Key },
		)
		u.ProfilesToDelete = namesToDelete(p.Profiles, u.Profiles,
			func(m ClickHouseProfileModel) string { return m.Name.ValueString() },
			func(s *sdk.ClickHouseProfileUpdateSpecInput) string { return s.Name },
		)
		u.UsersToDelete = namesToDelete(p.Users, u.Users,
			func(m ClickHouseUserModel) string { return m.Name.ValueString() },
			func(s *sdk.ClickHouseUserSpecInput) string { return s.Name },
		)

		priorProfiles := make(map[string][]ClickHouseSettingModel, len(p.Profiles))
		for _, pp := range p.Profiles {
			priorProfiles[pp.Name.ValueString()] = pp.Settings
		}
		for _, up := range u.Profiles {
			up.SettingsToDelete = namesToDelete(priorProfiles[up.Name], up.Settings,
				func(m ClickHouseSettingModel) string { return m.Key.ValueString() },
				func(s *sdk.ClickHouseSettingSpecInput) string { return s.Key },
			)
		}
	}
}

func namesToDelete[P any, N any](prior []P, planned []N, priorKey func(P) string, plannedKey func(N) string) []string {
	kept := make(map[string]bool, len(planned))
	for _, n := range planned {
		kept[plannedKey(n)] = true
	}

	var deleted []string
	for _, p := range prior {
		if key := priorKey(p); !kept[key] {
			deleted = append(deleted, key)
		}
	}

	return deleted
}

func clickHouseDiskToSDK(name string, disk *ClickHouseDiskModel) *sdk.ClickHouseDiskCreateSpecInput {
	if disk == nil {
		return nil
	}

	return &sdk.ClickHouseDiskCreateSpecInput{
		Name:         name,
		Size:         knownInt64Pointer(disk.Size),
		StorageClass: knownStringPointer(disk.StorageClass),
		Iops:         knownInt64Pointer(disk.Iops),
		Throughput:   knownInt64Pointer(disk.Throughput),
	}
}

func clickHouseDiskToUpdateSDK(name string, disk *ClickHouseDiskModel) *sdk.ClickHouseDiskUpdateSpecInput {
	if disk == nil {
		return nil
	}

	return &sdk.ClickHouseDiskUpdateSpecInput{
		Name:       name,
		Size:       knownInt64Pointer(disk.Size),
		Iops:       knownInt64Pointer(disk.Iops),
		Throughput: knownInt64Pointer(disk.Throughput),
	}
}

func clickHouseKeeperRefToSDK(keeper *ClickHouseKeeperRefModel) *sdk.ClickHouseKeeperSpecInput {
	if keeper == nil {
		return nil
	}

	return &sdk.ClickHouseKeeperSpecInput{
		Enabled: keeper.Enabled.ValueBool(),
		Name:    keeper.Name.ValueString(),
	}
}

func clickHouseSettingsToSDK(settings []ClickHouseSettingModel) []*sdk.ClickHouseSettingSpecInput {
	var input []*sdk.ClickHouseSettingSpecInput
	for _, s := range settings {
		input = append(input, &sdk.ClickHouseSettingSpecInput{
			Key:             s.Key.ValueString(),
			Value:           knownStringPointer(s.Value),
			ValueFromSecret: clickHouseSecretRefToSDK(s.ValueFromSecret),
		})
	}

	return input
}

func clickHouseProfilesToSDK(profiles []ClickHouseProfileModel) []*sdk.ClickHouseProfileCreateSpecInput {
	var input []*sdk.ClickHouseProfileCreateSpecInput
	for _, p := range profiles {
		input = append(input, &sdk.ClickHouseProfileCreateSpecInput{
			Name:     p.Name.ValueString(),
			Settings: clickHouseSettingsToSDK(p.Settings),
		})
	}

	return input
}

func clickHouseProfilesToUpdateSDK(profiles []ClickHouseProfileModel) []*sdk.ClickHouseProfileUpdateSpecInput {
	var input []*sdk.ClickHouseProfileUpdateSpecInput
	for _, p := range profiles {
		input = append(input, &sdk.ClickHouseProfileUpdateSpecInput{
			Name:     p.Name.ValueString(),
			Settings: clickHouseSettingsToSDK(p.Settings),
		})
	}

	return input
}

func clickHouseUsersToSDK(ctx context.Context, users []ClickHouseUserModel) ([]*sdk.ClickHouseUserSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var input []*sdk.ClickHouseUserSpecInput

	for _, u := range users {
		allowedCIDRs, diags := stringListToSDK(ctx, u.AllowedCIDRs)
		allDiags.Append(diags...)
		databases, diags := stringListToSDK(ctx, u.Databases)
		allDiags.Append(diags...)

		input = append(input, &sdk.ClickHouseUserSpecInput{
			Name:                        u.Name.ValueString(),
			Profile:                     knownStringPointer(u.Profile),
			Quota:                       knownStringPointer(u.Quota),
			AllowedCIDRs:                allowedCIDRs,
			Databases:                   databases,
			AccessManagement:            knownBoolPointer(u.AccessManagement),
			NamedCollectionControl:      knownBoolPointer(u.NamedCollectionControl),
			ShowNamedCollections:        knownBoolPointer(u.ShowNamedCollections),
			ShowNamedCollectionsSecrets: knownBoolPointer(u.ShowNamedCollectionsSecrets),
			PasswordType:                (*sdk.ClickHouseUserPasswordTypeSpecInput)(knownStringPointer(u.PasswordType)),
			PasswordValue:               knownStringPointer(u.PasswordValue),
			PasswordValueFromSecret:     clickHouseSecretRefToSDK(u.PasswordValueFromSecret),
		})
	}

	return input, allDiags
}

func clickHouseSecretRefToSDK(ref *ClickHouseSecretRefModel) *sdk.ClickHouseSecretRefSpecInput {
	if ref == nil {
		return nil
	}

	return &sdk.ClickHouseSecretRefSpecInput{
		Name: ref.Name.ValueString(),
		Key:  ref.Key.ValueString(),
	}
}

func stringListToSDK(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	return values, diags
}

// A pointer method maps only null to nil: on unknown it returns a pointer to "" or 0.
func knownStringPointer(value types.String) *string {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

func knownInt64Pointer(value types.Int64) *int64 {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueInt64Pointer()
}

func knownBoolPointer(value types.Bool) *bool {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueBoolPointer()
}
