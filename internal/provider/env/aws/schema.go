package env

import (
	"context"
	"regexp"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *AWSEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) AWS environment resource.`),
		Attributes: map[string]rschema.Attribute{
			"id":                              common.IDAttribute,
			"name":                            common.NameAttribute,
			"custom_domain":                   common.GetCommonCustomDomainAttribute(false, true, false),
			"load_balancers":                  getLoadBalancersAttribute(false, true, true),
			"load_balancing_strategy":         common.GetLoadBalancingStrategyAttribute(false, true, true),
			"maintenance_windows":             common.GetMaintenanceWindowAttribute(false, true, false),
			"cidr":                            common.GetCIDRAttribute(true, false, false),
			"zones":                           getZonesAttribute(false, true, true, common.AWS_ZONES_DESCRIPTION),
			"node_groups":                     common.GetNodeGroupsAttribute(true, false, false),
			"aws_account_id":                  getAWSAccountIDAttribute(true, false, false),
			"region":                          common.GetRegionAttribute(true, false, false, common.AWS_REGION_DESCRIPTION),
			"nat":                             getNATAttribute(false, true, true),
			"peering_connections":             getPeeringConnectionsAttribute(false, true, false),
			"endpoints":                       getEndpointsAttribute(false, true, false),
			"tags":                            getTagsAttribute(false, true, false),
			"cloud_connect":                   getCloudConnectAttribute(false, true, true),
			"spec_revision":                   common.SpecRevisionAttribute,
			"force_destroy":                   common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, true, true),
		},
	}
}

func (d *AWSEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) AWS environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                      common.IDAttribute,
			"name":                    common.NameAttribute,
			"custom_domain":           common.GetCommonCustomDomainAttribute(false, false, true),
			"load_balancers":          getLoadBalancersAttribute(false, false, true),
			"load_balancing_strategy": common.GetLoadBalancingStrategyAttribute(false, false, true),
			"maintenance_windows":     common.GetMaintenanceWindowAttribute(false, false, true),
			"cidr":                    common.GetCIDRAttribute(false, false, true),
			"zones":                   getZonesAttribute(false, false, true, common.AWS_ZONES_DESCRIPTION),
			"node_groups":             common.GetNodeGroupsAttribute(false, false, true),
			"aws_account_id":          getAWSAccountIDAttribute(false, false, true),
			"region":                  common.GetRegionAttribute(false, false, true, common.AWS_REGION_DESCRIPTION),
			"nat":                     getNATAttribute(false, true, true),
			"peering_connections":     getPeeringConnectionsAttribute(false, false, true),
			"endpoints":               getEndpointsAttribute(false, false, true),
			"tags":                    getTagsAttribute(false, false, true),
			"cloud_connect":           getCloudConnectAttribute(false, false, true),
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

func getLoadBalancersAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.LOAD_BALANCER_DESCRIPTION,
		PlanModifiers: []planmodifier.Object{
			modifiers.DefaultObject(map[string]attr.Value{
				"public":   loadBalancerPublicDefaultObject,
				"internal": loadBalancerInternalDefaultObject,
			}),
		},
		Attributes: map[string]rschema.Attribute{
			"public": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(loadBalancerPublicDefaultObject),
				MarkdownDescription: common.LOAD_BALANCER_PUBLIC_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
					"cross_zone":       crossZoneAttribute,
				},
			},
			"internal": rschema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(loadBalancerInternalDefaultObject),
				MarkdownDescription: common.LOAD_BALANCER_INTERNAL_DESCRIPTION,
				Attributes: map[string]rschema.Attribute{
					"enabled":          common.EnabledAttribute,
					"source_ip_ranges": common.SourceIPRangesAttribute,
					"cross_zone":       crossZoneAttribute,
					"endpoint_service_allowed_principals": rschema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: common.AWS_LOAD_BALANCER_ENDPOINT_SERVICE_ALLOWED_PRINCIPALS_DESCRIPTION,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
				},
			},
		},
	}
}

func getAWSAccountIDAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.AWS_ACCOUNT_ID_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("aws_account_id"),
		},
		Validators: []validator.String{
			stringvalidator.RegexMatches(regexp.MustCompile(`^\d{12}$`),
				"must be a 12-digit number"),
		},
	}
}

func getPeeringConnectionsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		NestedObject:        peeringAttribute,
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.PEERING_CONNECTION_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

func getTagsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return common.GetTagsAttribute(required, optional, computed, common.AWS_TAGS_DESCRIPTION)
}

func getEndpointsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		NestedObject:        endpointAttribute,
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.ENDPOINT_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

func getCloudConnectAttribute(required, optional, computed bool) rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.CLOUD_CONNECT_DESCRIPTION,
		Default:             booldefault.StaticBool(true),
	}
}

func getZonesAttribute(required, optional, computed bool, description string) rschema.ListAttribute {
	zonesAttribute := common.GetZonesAttribute(required, optional, computed, description)
	zonesAttribute.Validators = []validator.List{
		listvalidator.SizeAtLeast(2),
	}

	return zonesAttribute
}

func getNATAttribute(required, optional, computed bool) rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Required:            required,
		Optional:            optional,
		Computed:            computed,
		MarkdownDescription: common.NAT_DESCRIPTION,
		Default:             booldefault.StaticBool(false),
	}
}

var endpointAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"service_name": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.ENDPOINT_SERVICE_NAME_DESCRIPTION,
		},
		"alias": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.ENDPOINT_ALIAS_DESCRIPTION,
		},
	},
}

var peeringAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"aws_account_id": rschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: common.AWS_ACCOUNT_ID_DESCRIPTION,
		},
		"vpc_id": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.PEERING_CONNECTION_VPC_ID_DESCRIPTION,
		},
		"vpc_region": rschema.StringAttribute{
			Optional:            true,
			MarkdownDescription: common.PEERING_CONNECTION_VPC_REGION_DESCRIPTION,
		},
	},
}

var crossZoneAttribute = rschema.BoolAttribute{
	Optional:            true,
	Computed:            true,
	MarkdownDescription: common.AWS_LOAD_BALANCER_CROSS_ZONE_DESCRIPTION,
	Default:             booldefault.StaticBool(false),
}

var loadBalancerInternalDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled":    types.BoolType,
		"cross_zone": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
		"endpoint_service_allowed_principals": types.ListType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"enabled":                             types.BoolValue(false),
		"cross_zone":                          types.BoolValue(false),
		"source_ip_ranges":                    types.ListNull(types.StringType),
		"endpoint_service_allowed_principals": types.ListNull(types.StringType),
	},
)

var loadBalancerPublicDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled":    types.BoolType,
		"cross_zone": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"enabled":          types.BoolValue(false),
		"cross_zone":       types.BoolValue(false),
		"source_ip_ranges": types.ListNull(types.StringType),
	},
)
