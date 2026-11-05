package env

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KeyValueModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type NodeGroupsModel struct {
	Name            types.String `tfsdk:"name"`
	NodeType        types.String `tfsdk:"node_type"`
	CapacityPerZone types.Int64  `tfsdk:"capacity_per_zone"`
	Zones           types.List   `tfsdk:"zones"`
	Reservations    types.Set    `tfsdk:"reservations"`
}

type MetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

type DatadogModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	EncAPIKey      types.String `tfsdk:"enc_api_key"`
	Domain         types.String `tfsdk:"domain"`
	LogsEnabled    types.Bool   `tfsdk:"logs_enabled"`
	MetricsEnabled types.Bool   `tfsdk:"metrics_enabled"`
}

type MaintenanceWindowModel struct {
	Name          types.String   `tfsdk:"name"`
	Enabled       types.Bool     `tfsdk:"enabled"`
	Hour          types.Int64    `tfsdk:"hour"`
	LengthInHours types.Int64    `tfsdk:"length_in_hours"`
	Days          []types.String `tfsdk:"days"`
}

// Config order, paired by key; unmatched items last, duplicate keys pair positionally, nil stays nil.
func ReorderByKey[M any, S any](model []M, items []S, getModelKey func(M) string, getItemKey func(S) string) []S {
	if len(items) == 0 {
		return items
	}

	ordered := make([]S, 0, len(items))
	used := make([]bool, len(items))

	for _, m := range model {
		mk := getModelKey(m)
		for i, item := range items {
			if !used[i] && mk == getItemKey(item) {
				ordered = append(ordered, item)
				used[i] = true
				break
			}
		}
	}

	for i, item := range items {
		if !used[i] {
			ordered = append(ordered, item)
		}
	}

	return ordered
}

// Pairing mirrors ReorderByKey so both resolve the same model/item pairs; mutates items via itemList.
func ReorderNodeGroupZones[M any, S any](
	ctx context.Context,
	model []M,
	items []S,
	getModelKey func(M) string,
	getItemKey func(S) string,
	modelList func(M) types.List,
	itemList func(S) *[]string,
) diag.Diagnostics {
	var allDiags diag.Diagnostics
	used := make([]bool, len(items))

	for _, m := range model {
		mk := getModelKey(m)
		for i, item := range items {
			if !used[i] && mk == getItemKey(item) {
				used[i] = true
				target := itemList(item)
				reordered, diags := ReorderList(ctx, modelList(m), *target)
				allDiags.Append(diags...)
				if !diags.HasError() {
					*target = reordered
				}
				break
			}
		}
	}

	return allDiags
}

func ReorderList(ctx context.Context, model types.List, input []string) ([]string, diag.Diagnostics) {
	if model.IsUnknown() || model.IsNull() {
		return input, nil
	}

	var modelValues []string
	diags := model.ElementsAs(ctx, &modelValues, false)
	if diags.HasError() {
		return nil, diags
	}

	return reorderStrings(modelValues, input), diags
}

// Model order first, then unmatched input in API order; set-keyed, so a repeated input value is kept once.
func reorderStrings(model, input []string) []string {
	ordered := make([]string, 0, len(input))
	used := make(map[string]bool, len(input))

	for _, m := range model {
		for _, v := range input {
			if m == v {
				ordered = append(ordered, v)
				used[v] = true
				break
			}
		}
	}

	for _, v := range input {
		if !used[v] {
			ordered = append(ordered, v)
		}
	}

	return ordered
}
