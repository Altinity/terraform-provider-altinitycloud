package env

import (
	"context"

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

type MaintenanceWindowModel struct {
	Name          types.String   `tfsdk:"name"`
	Enabled       types.Bool     `tfsdk:"enabled"`
	Hour          types.Int64    `tfsdk:"hour"`
	LengthInHours types.Int64    `tfsdk:"length_in_hours"`
	Days          []types.String `tfsdk:"days"`
}

func ReorderList(model types.List, input []string) []string {
	orderedZones := make([]string, 0, len(input))

	var modelZones []string
	model.ElementsAs(context.TODO(), &modelZones, false)

	for _, zone := range modelZones {
		for _, apiZone := range input {
			if zone == apiZone {
				orderedZones = append(orderedZones, apiZone)
				break
			}
		}
	}

	return orderedZones
}
