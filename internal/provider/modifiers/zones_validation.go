package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ZonesAttributePlanModifier() planmodifier.Int64 {
	return &zonesAttributePlanModifier{}
}

type zonesAttributePlanModifier struct {
}

func (d *zonesAttributePlanModifier) Description(ctx context.Context) string {
	return "Ensures that attribute_one and attribute_two attributes are kept synchronised."
}

func (d *zonesAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *zonesAttributePlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var zones types.List
	diags := req.Plan.GetAttribute(ctx, path.Root("zones"), &zones)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var numberOfZones types.Int64
	req.Plan.GetAttribute(ctx, path.Root("number_of_zones"), &numberOfZones)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if zones.IsNull() && numberOfZones.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("zones"), "zones or number_of_zones are required", "you must set one of them")
		return
	}

	// if !zones.IsNull() && numberOfZones.ValueInt64() > 0 {
	// 	resp.Diagnostics.AddAttributeWarning(path.Root("zones"), "zones and number_of_zones are mutually exclusive", "if you set both, only zones will be used")
	// }

	if !zones.IsNull() && len(zones.Elements()) > 0 {
		resp.PlanValue = types.Int64Value(int64(len(zones.Elements())))
		return
	}
}
