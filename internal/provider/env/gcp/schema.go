package env

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *GCPEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) GCP environment resource.`),
		Attributes: map[string]rschema.Attribute{
			"id":                              common.IDAttribute,
			"name":                            common.NameAttribute,
			"custom_domain":                   common.GetCommonCustomDomainAttribute(false, true, false),
			"load_balancers":                  getLoadBalancersAttribute(false, true, true),
			"load_balancing_strategy":         common.GetLoadBalancingStrategyAttribute(false, true, true),
			"maintenance_windows":             common.GetMaintenanceWindowAttribute(false, true, false),
			"cidr":                            common.GetCIDRAttribute(true, false, false),
			"zones":                           common.GetZonesAttribute(false, true, true, common.GCP_ZONES_DESCRIPTION),
			"node_groups":                     common.GetNodeGroupsAttribute(true, false, false),
			"region":                          common.GetRegionAttribute(true, false, false, common.GCP_REGION_DESCRIPTION),
			"gcp_project_id":                  getGCPProjectIDAttribute(true, false, false),
			"spec_revision":                   common.SpecRevisionAttribute,
			"force_destroy":                   common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, true, true),
		},
	}
}

func (d *GCPEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) GCP environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                      common.IDAttribute,
			"name":                    common.NameAttribute,
			"custom_domain":           common.GetCommonCustomDomainAttribute(false, false, true),
			"load_balancers":          getLoadBalancersAttribute(false, false, true),
			"load_balancing_strategy": common.GetLoadBalancingStrategyAttribute(false, false, true),
			"maintenance_windows":     common.GetMaintenanceWindowAttribute(false, false, true),
			"cidr":                    common.GetCIDRAttribute(false, false, true),
			"zones":                   common.GetZonesAttribute(false, false, true, common.GCP_ZONES_DESCRIPTION),
			"node_groups":             common.GetNodeGroupsAttribute(false, false, true),
			"gcp_project_id":          getGCPProjectIDAttribute(false, false, true),
			"region":                  common.GetRegionAttribute(false, false, true, common.GCP_REGION_DESCRIPTION),
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
					"source_ip_ranges": common.SourceIPRangesAttribute,
				},
			},
		},
	}
}

func getGCPProjectIDAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.GCP_PROJECT_ID_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("gcp_project_id"),
		},
	}
}

var loadBalancerDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"enabled": types.BoolType,
		"source_ip_ranges": types.ListType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"enabled":          types.BoolValue(false),
		"source_ip_ranges": types.ListNull(types.StringType),
	},
)
