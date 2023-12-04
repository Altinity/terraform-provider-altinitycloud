package env

import (
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
