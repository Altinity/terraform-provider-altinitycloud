package env

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *K8SEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Kubernetes (BYOK) environment resource.`),
		Attributes: map[string]rschema.Attribute{
			"id":                              common.IDAttribute,
			"name":                            common.NameAttribute,
			"custom_domain":                   common.GetCommonCustomDomainAttribute(false, true, false),
			"load_balancers":                  getLoadBalancersAttribute(false, true, true),
			"load_balancing_strategy":         common.GetLoadBalancingStrategyAttribute(false, true, true),
			"maintenance_windows":             common.GetMaintenanceWindowAttribute(false, true, false),
			"logs":                            getLogsAttributes(false, true, true),
			"metrics":                         getMetricsAttribute(false, true, true),
			"distribution":                    getDistributionAttribute(true, false, false),
			"node_groups":                     getNodeGroupsAttribute(true, false, false),
			"custom_node_types":               getCustomNodeTypes(false, true, false),
			// "metrics_endpoint":                common.GetMetricsEndpointAttribute(false, true, false),
			"spec_revision":                   common.SpecRevisionAttribute,
			"force_destroy":                   common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, true, true),
		},
	}
}

func (d *K8SEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Kubernetes (BYOK) environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                      common.IDAttribute,
			"name":                    common.NameAttribute,
			"custom_domain":           common.GetCommonCustomDomainAttribute(false, false, true),
			"load_balancers":          getLoadBalancersAttribute(false, false, true),
			"load_balancing_strategy": common.GetLoadBalancingStrategyAttribute(false, false, true),
			"maintenance_windows":     common.GetMaintenanceWindowAttribute(false, false, true),
			"logs":                    getLogsAttributes(false, false, true),
			"metrics":                 getMetricsAttribute(false, false, true),
			"distribution":            getDistributionAttribute(false, false, true),
			"node_groups":             getNodeGroupsAttribute(false, false, true),
			"custom_node_types":       getCustomNodeTypes(false, false, true),
			// "metrics_endpoint":        common.GetMetricsEndpointAttribute(false, false, true),
			"spec_revision":           common.SpecRevisionAttribute,

			// these options are not used in data sources,
			// but we need to include them in the schema to avoid conversion errors.
			"force_destroy":                   common.GetForceDestroyAttribute(false, false, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, false, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, false, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, false, true),
		},
	}
}

func getDistributionAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.DISTRIBUTION_DESCRIPTION,
		Validators: []validator.String{
			stringvalidator.OneOf([]string{
				string(client.K8SDistributionCustom),
				string(client.K8SDistributionAks),
				string(client.K8SDistributionEks),
				string(client.K8SDistributionGke)}...,
			),
		},
	}
}

func getLogsAttributes(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Attributes:          storageAttribute,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.LOGS_DESCRIPTION,
		PlanModifiers: []planmodifier.Object{
			modifiers.DefaultObject(map[string]attr.Value{
				"storage": storagDefaultValue,
			}),
		},
	}
}

func getNodeGroupsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		NestedObject:        nodeGroupAttribute,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.NODE_GROUP_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

func getCustomNodeTypes(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		NestedObject:        nodeTypeAttribute,
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.CUSTOM_NODE_TYPES_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

func getMetricsAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.METRICS_DESCRIPTION,
		PlanModifiers: []planmodifier.Object{
			modifiers.DefaultObject(map[string]attr.Value{
				"retention_period_in_days": types.Int64Null(),
			}),
		},
		Attributes: map[string]rschema.Attribute{
			"retention_period_in_days": rschema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: common.METRICS_RETENTION_PERIOD_IN_DAYS_DESCRIPTION,
			},
		},
	}
}

func getLoadBalancersAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.LOAD_BALANCER_DESCRIPTION,
		PlanModifiers: []planmodifier.Object{
			modifiers.DefaultObject(map[string]attr.Value{
				"public":   loadBalancerDefaultObject,
				"internal": loadBalancerDefaultObject,
			}),
		},
		Attributes: map[string]rschema.Attribute{
			"public": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: common.LOAD_BALANCER_PUBLIC_DESCRIPTION,
				Default:             objectdefault.StaticValue(loadBalancerDefaultObject),
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"annotations":      annotationsAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
				},
			},
			"internal": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: common.LOAD_BALANCER_INTERNAL_DESCRIPTION,
				Default:             objectdefault.StaticValue(loadBalancerDefaultObject),
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"annotations":      annotationsAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
				},
			},
		},
	}
}

var annotationsAttribute = rschema.ListNestedAttribute{
	NestedObject:        common.KeyValueAttribute,
	Optional:            true,
	MarkdownDescription: common.K8S_LOAD_BALANCER_ANNOTATIONS_DESCRIPTION,
	Validators: []validator.List{
		listvalidator.SizeAtLeast(1),
	},
}

var storageAttribute = map[string]rschema.Attribute{
	"storage": rschema.SingleNestedAttribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: common.STORAGE_DESCRIPTION,
		Default:             objectdefault.StaticValue(storagDefaultValue),
		Attributes: map[string]rschema.Attribute{
			"s3": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(s3DefaultValue),
				MarkdownDescription: common.S3_STORAGE_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"bucket_name": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.BUCKET_NAME_DESCRIPTION,
					},
					"region": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.AWS_REGION_DESCRIPTION,
					},
				},
			},
			"gcs": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(gcsDefaultValue),
				MarkdownDescription: common.GCS_STORAGE_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"bucket_name": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: common.BUCKET_NAME_DESCRIPTION,
					},
				},
			},
		},
	},
}

var nodeGroupAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"name": rschema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: common.NODE_GROUP_NAME_DESCRIPTION,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"node_type": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.K8S_NODE_GROUP_NODE_TYPE_DESCRIPTION,
		},
		"capacity_per_zone": rschema.Int64Attribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_CAPACITY_PER_ZONE_DESCRIPTION,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"tolerations": rschema.ListNestedAttribute{
			NestedObject:        tolerationsAttribute,
			Optional:            true,
			MarkdownDescription: common.NODE_GROUP_TOLERATIONS,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"selector": rschema.ListNestedAttribute{
			NestedObject:        common.KeyValueAttribute,
			Optional:            true,
			MarkdownDescription: common.NODE_GROUP_SELECTOR_DESCRIPTION,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"zones": rschema.ListAttribute{
			ElementType:         types.StringType,
			Required:            true,
			MarkdownDescription: common.K8S_NODE_GROUP_ZONES_DESCRIPTION,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"reservations": common.GetReservationsAttribute(false, true, true),
	},
}

// TODO: Add description.
var tolerationsAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"key": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_TOLERATIONS_KEY,
		},
		"value": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_TOLERATIONS_VALUE,
		},
		"operator": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_TOLERATIONS_OPERATOR,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(client.NodeTolerationOperatorExists),
					string(client.NodeTolerationOperatorEqual)}...,
				),
			},
		},
		"effect": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_TOLERATIONS_EFFECT,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(client.NodeTolerationEffectNoSchedule),
					string(client.NodeTolerationEffectPreferNoSchedule),
					string(client.NodeTolerationEffectNoExecute)}...,
				),
			},
		},
	},
}

var nodeTypeAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"name": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.CUSTOM_NODE_TYPES_NAME_DESCRIPTION,
		},
		"cpu_allocatable": rschema.Float64Attribute{
			Optional:            true,
			MarkdownDescription: common.CUSTOM_NODE_TYPES_CPU_ALLOCATABLE_DESCRIPTION,
			Validators: []validator.Float64{
				float64validator.AtLeast(0),
			},
		},
		"mem_allocatable_in_bytes": rschema.Float64Attribute{
			Optional:            true,
			MarkdownDescription: common.CUSTOM_NODE_TYPES_MEMORY_ALLOCATABLE__DESCRIPTION,
			Validators: []validator.Float64{
				float64validator.AtLeast(0),
			},
		},
	},
}

var gcsDefaultValue = types.ObjectNull(
	map[string]attr.Type{
		"bucket_name": types.StringType,
	},
)

var storageDefaultType = map[string]attr.Type{
	"s3": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"bucket_name": types.StringType,
			"region":      types.StringType,
		},
	},
	"gcs": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"bucket_name": types.StringType,
		},
	},
}

var s3DefaultValue = types.ObjectNull(
	map[string]attr.Type{
		"bucket_name": types.StringType,
		"region":      types.StringType,
	},
)

var storagDefaultValue = types.ObjectValueMust(
	storageDefaultType,
	map[string]attr.Value{
		"s3":  s3DefaultValue,
		"gcs": gcsDefaultValue,
	},
)

var loadBalancerDefaultObject = types.ObjectValueMust(
	map[string]attr.Type{
		"enabled": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
		"annotations": types.ListType{
			ElemType: annotationsDefaultType,
		},
	},
	map[string]attr.Value{
		"enabled":          types.BoolValue(false),
		"source_ip_ranges": types.ListNull(types.StringType),
		"annotations":      types.ListNull(annotationsDefaultType),
	},
)

var annotationsDefaultType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	},
}
