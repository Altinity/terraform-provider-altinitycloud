package env

import (
	"context"
	"maps"
	"slices"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Quotas are the one ClickHouse input the API does not split per env.
func ClickHouseQuotasToSDK(quotas map[string]ClickHouseQuotaModel) []*sdk.ClickHouseQuotaSpecInput {
	var input []*sdk.ClickHouseQuotaSpecInput

	for _, name := range SortedKeys(quotas) {
		q := quotas[name]
		var intervals []*sdk.ClickHouseQuotaIntervalSpecInput
		for _, i := range q.Intervals {
			intervals = append(intervals, &sdk.ClickHouseQuotaIntervalSpecInput{
				Duration:                        uint64(i.DurationSeconds.ValueInt64()),
				Queries:                         knownUint64Pointer(i.Queries),
				QuerySelects:                    knownUint64Pointer(i.QuerySelects),
				QueryInserts:                    knownUint64Pointer(i.QueryInserts),
				Errors:                          knownUint64Pointer(i.Errors),
				ResultRows:                      knownUint64Pointer(i.ResultRows),
				ResultBytes:                     knownUint64Pointer(i.ResultBytes),
				ReadRows:                        knownUint64Pointer(i.ReadRows),
				ReadBytes:                       knownUint64Pointer(i.ReadBytes),
				WrittenBytes:                    knownUint64Pointer(i.WrittenBytes),
				ExecutionTime:                   knownFloat64Pointer(i.ExecutionTimeSeconds),
				FailedSequentialAuthentications: knownUint64Pointer(i.FailedSequentialAuthentications),
				QueriesPerNormalizedHash:        knownUint64Pointer(i.QueriesPerNormalizedHash),
			})
		}

		var keyedBy *sdk.ClickHouseQuotaKeyedBy
		if value, ok := knownString(q.KeyedBy); ok {
			k := sdk.ClickHouseQuotaKeyedBy(value)
			keyedBy = &k
		}

		input = append(input, &sdk.ClickHouseQuotaSpecInput{
			Name:           name,
			KeyedBy:        keyedBy,
			Ipv4PrefixBits: KnownInt64Pointer(q.IPv4PrefixBits),
			Ipv6PrefixBits: KnownInt64Pointer(q.IPv6PrefixBits),
			Intervals:      intervals,
		})
	}

	return input
}

func StringListToSDK(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	return values, diags
}

// A pointer method maps only null to nil: on unknown it returns a pointer to "" or 0.
func KnownStringPointer(value types.String) *string {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

func KnownInt64Pointer(value types.Int64) *int64 {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueInt64Pointer()
}

func KnownBoolPointer(value types.Bool) *bool {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueBoolPointer()
}

func knownFloat64Pointer(value types.Float64) *float64 {
	if value.IsUnknown() {
		return nil
	}
	return value.ValueFloat64Pointer()
}

// Terraform has no unsigned type, so quota limits ride Int64 with a non-negative validator.
func knownUint64Pointer(value types.Int64) *uint64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	u := uint64(value.ValueInt64())
	return &u
}

func knownString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", false
	}
	return value.ValueString(), true
}

// Map iteration is random; a stable order keeps requests and tests deterministic.
func SortedKeys[M any](m map[string]M) []string {
	return slices.Sorted(maps.Keys(m))
}

// A patch only touches what Terraform owns, so a dropped entry must be named or the API keeps it.
func ClickHouseNamesToDelete[M any](prior, planned map[string]M) []string {
	var deleted []string
	for _, name := range SortedKeys(prior) {
		if _, ok := planned[name]; !ok {
			deleted = append(deleted, name)
		}
	}

	return deleted
}

// Patch Replace upserts by name, so an added name colliding with an unmanaged entry would overwrite it.
func ClickHouseNamesAdded[M any](prior, planned map[string]M) []string {
	return ClickHouseNamesToDelete(planned, prior)
}

func ClickHouseCollisions(added, existing []string) []string {
	existingSet := make(map[string]bool, len(existing))
	for _, e := range existing {
		existingSet[e] = true
	}

	var collisions []string
	for _, a := range added {
		if existingSet[a] {
			collisions = append(collisions, a)
		}
	}

	return collisions
}
