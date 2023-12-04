package common

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var CIDR_REGEX = regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}\/[0-9]{1,2}$`)
var DOMAIN_REGEX = regexp.MustCompile("^[a-z0-9][a-z0-9-]{0,63}([.][a-z0-9][a-z0-9-]{0,63})+$")

func GetCustomDomainAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: CUSTOM_DOMAIN_DESCRIPTION,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				DOMAIN_REGEX,
				"invalid domain format",
			),
		},
	}
}

func GetMaintenanceWindowAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: MAINTENANCE_WINDOW_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: rschema.NestedAttributeObject{
			Attributes: map[string]rschema.Attribute{
				"name": rschema.StringAttribute{
					Required:            true,
					MarkdownDescription: MAINTENANCE_WINDOW_NAME_DESCRIPTION,
				},
				"enabled": rschema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: MAINTENANCE_WINDOW_ENABLED_DESCRIPTION,
					Default:             booldefault.StaticBool(false),
				},
				"hour": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: MAINTENANCE_WINDOW_HOUR_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
						int64validator.AtMost(23),
					},
				},
				"length_in_hours": rschema.Int64Attribute{
					Required:            true,
					MarkdownDescription: MAINTENANCE_WINDOW_LENGTH_IN_HOURS_DESCRIPTION,
					Validators: []validator.Int64{
						int64validator.AtLeast(4),
						int64validator.AtMost(24),
					},
				},
				"days": rschema.ListAttribute{
					ElementType:         types.StringType,
					Required:            true,
					MarkdownDescription: MAINTENANCE_WINDOW_DAYS_DESCRIPTION,
					Validators: []validator.List{
						listvalidator.ValueStringsAre(
							stringvalidator.OneOf([]string{
								string(client.DayMonday),
								string(client.DayTuesday),
								string(client.DayWednesday),
								string(client.DayThursday),
								string(client.DayFriday),
								string(client.DaySaturday),
								string(client.DaySunday)}...,
							),
						),
					},
				},
			},
		},
	}
}

func GetCIDRAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: CIDR_DESCRIPTION,
		Validators: []validator.String{
			stringvalidator.RegexMatches(
				CIDR_REGEX,
				"invalid CIDR format ([IP Address]/[Prefix Length])",
			),
		},
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("cidr"),
		},
	}
}

func GetZonesAttribute(required, optional, computed bool, description string) rschema.ListAttribute {
	return rschema.ListAttribute{
		ElementType:         types.StringType,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: description,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

func GetNumberOfZonesAttribute(required, optional, computed bool) rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: NUMBER_OF_ZONES_DESCRIPTION,
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
		},
		PlanModifiers: []planmodifier.Int64{
			modifiers.ZonesAttributePlanModifier(),
		},
	}
}

func GetLoadBalancingStrategyAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		Default:             stringdefault.StaticString(string(client.LoadBalancingStrategyZoneBestEffort)),
		MarkdownDescription: LOAD_BALANCING_STRATEGY_DESCRIPTION,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{
				string(client.LoadBalancingStrategyRoundRobin),
				string(client.LoadBalancingStrategyZoneBestEffort)}...,
			),
		},
	}
}

func GetNodeGroupsAttribure(required, optional, computed bool) rschema.SetNestedAttribute {
	return rschema.SetNestedAttribute{
		NestedObject:        NodeGroupAttribute,
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: NODE_GROUP_DESCRIPTION,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	}
}

func GetRegionAttribure(required, optional, computed bool, description string) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: description,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("region"),
		},
	}
}

func GetForceDestroyAttribute(required, optional, computed bool) rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: FORCE_DESTROY_DESCRIPTION,
		Default:             booldefault.StaticBool(false),
	}
}

func GetSkipProvisioningOnDestroyAttribute(required, optional, computed bool) rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Required: required,
		Optional: optional,
		Computed: computed,
		Default:  booldefault.StaticBool(false),
	}
}

func GetReservationsAttribute(required, optional, computed bool) rschema.SetAttribute {
	return rschema.SetAttribute{
		ElementType:         types.StringType,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: NODE_GROUP_RESERVATIONS_DESCRIPTION,
		PlanModifiers: []planmodifier.Set{
			modifiers.DefaultSet(types.StringType, []attr.Value{
				types.StringValue(string(client.NodeReservationClickhouse)),
				types.StringValue(string(client.NodeReservationSystem)),
				types.StringValue(string(client.NodeReservationZookeeper)),
			}),
		},
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
			setvalidator.ValueStringsAre(
				stringvalidator.OneOf([]string{
					string(client.NodeReservationClickhouse),
					string(client.NodeReservationSystem),
					string(client.NodeReservationZookeeper)}...,
				),
			),
		},
	}
}

var PendingDeleteAttribute = rschema.BoolAttribute{
	Required:            false,
	Optional:            false,
	Computed:            true,
	MarkdownDescription: STATUS_PENDING_DELETE_DESCRIPTION,
}

var SpecRevisionAttribute = rschema.Int64Attribute{
	Computed:            true,
	MarkdownDescription: STATUS_SPEC_REVISION_DESCRIPTION,
}

var WaitForAppliedSpecRevisionAttribute = rschema.Int64Attribute{
	Optional:            true,
	MarkdownDescription: STATUS_APPLIED_SPEC_REVISION_DESCRIPTION,
}

var AppliedSpecRevisionAttribute = rschema.Int64Attribute{
	Computed:            true,
	MarkdownDescription: STATUS_APPLIED_SPEC_REVISION_DESCRIPTION,
}

var IDAttribute = rschema.StringAttribute{
	Computed:            true,
	MarkdownDescription: ID_DESCRIPTION,
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.UseStateForUnknown(),
	},
}

var NameAttribute = rschema.StringAttribute{
	Required:            true,
	MarkdownDescription: NAME_DESCRIPTION,
	PlanModifiers: []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	},
}

var SourceIPRangesAttribute = rschema.ListAttribute{
	ElementType:         types.StringType,
	Required:            true,
	MarkdownDescription: SOURCE_IP_RANGES_DESCRIPTION,
	Validators: []validator.List{
		listvalidator.SizeAtLeast(1),
		listvalidator.ValueStringsAre(
			stringvalidator.RegexMatches(CIDR_REGEX, "invalid CIDR (expecting something like 1.2.3.4/32)"),
		),
	},
}

var KeyValueAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: KEY_DESCRIPTION,
		},
		"value": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: VALUE_DESCRIPTION,
		},
	},
}

var NodeGroupAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"name": rschema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: NODE_GROUP_NAME_DESCRIPTION,
		},
		"node_type": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: NODE_GROUP_DESCRIPTION,
		},
		"capacity_per_zone": rschema.Int64Attribute{
			Required:            true,
			MarkdownDescription: NODE_GROUP_CAPACITY_PER_ZONE_DESCRIPTION,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"zones": rschema.ListAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			MarkdownDescription: NODE_GROUP_ZONES_DESCRIPTION,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"reservations": GetReservationsAttribute(true, false, false),
	},
}

var EnabledAttribute = rschema.BoolAttribute{
	Optional:            true,
	Computed:            true,
	MarkdownDescription: LOAD_BALANCER_ENABLED_DESCRIPTION,
	Default:             booldefault.StaticBool(false),
}
