package hosted_env

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodeGroupsModel struct {
	Name            types.String `tfsdk:"name"`
	NodeType        types.String `tfsdk:"node_type"`
	CapacityPerZone types.Int64  `tfsdk:"capacity_per_zone"`
	ZoneIDs         types.List   `tfsdk:"zone_ids"`
	Reservations    types.Set    `tfsdk:"reservations"`
}

type MetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}
