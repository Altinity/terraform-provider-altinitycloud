package env

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

func (r *AzureEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) Azure environment resource.`),
		Attributes: map[string]rschema.Attribute{
			"id":                          common.IDAttribute,
			"name":                        common.NameAttribute,
			"custom_domain":               getCustomDomainAttribute(false, true, false),
			"load_balancers":              getLoadBalancersAttribute(false, true, true),
			"load_balancing_strategy":     common.GetLoadBalancingStrategyAttribute(false, true, true),
			"maintenance_windows":         common.GetMaintenanceWindowAttribute(false, true, false),
			"cidr":                        common.GetCIDRAttribute(true, false, false),
			"zones":                       common.GetZonesAttribute(false, true, true, common.AZURE_ZONES_DESCRIPTION),
			"number_of_zones":             common.GetNumberOfZonesAttribute(false, true, true),
			"node_groups":                 common.GetNodeGroupsAttribure(true, false, false),
			"region":                      common.GetRegionAttribure(true, false, false, common.AZURE_REGION_DESCRIPTION),
			"tenant_id":                   getAzureTenantIDAttribute(true, false, false),
			"subscription_id":             getAzureSubscriptionIDAttribute(true, false, false),
			"tags":                        getTagsAttribute(false, true, false),
			"private_link_service":        getPrivateLinkServiceAttribute(false, true, true),
			"spec_revision":               common.SpecRevisionAttribute,
			"force_destroy":               common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":      common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy": common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"timeouts":                    common.GetTimeoutsAttribute(ctx),
		},
	}
}

func (d *AzureEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) Azure environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                      common.IDAttribute,
			"name":                    common.NameAttribute,
			"custom_domain":           getCustomDomainAttribute(false, false, true),
			"load_balancers":          getLoadBalancersAttribute(false, false, true),
			"load_balancing_strategy": common.GetLoadBalancingStrategyAttribute(false, false, true),
			"maintenance_windows":     common.GetMaintenanceWindowAttribute(false, false, true),
			"cidr":                    common.GetCIDRAttribute(false, false, true),
			"zones":                   common.GetZonesAttribute(false, false, true, common.AZURE_ZONES_DESCRIPTION),
			"number_of_zones":         common.GetNumberOfZonesAttribute(false, false, true),
			"node_groups":             common.GetNodeGroupsAttribure(false, false, true),
			"region":                  common.GetRegionAttribure(false, false, true, common.AZURE_REGION_DESCRIPTION),
			"tenant_id":               getAzureTenantIDAttribute(false, false, true),
			"subscription_id":         getAzureSubscriptionIDAttribute(false, false, true),
			"tags":                    getTagsAttribute(false, false, true),
			"private_link_service":    getPrivateLinkServiceAttribute(false, false, true),
			"spec_revision":           common.SpecRevisionAttribute,

			// these options are not used in data sources,
			// but we need to include them in the schema to avoid conversion errors.
			"force_destroy":               common.GetForceDestroyAttribute(false, false, true),
			"force_destroy_clusters":      common.GetForceDestroyClustersAttribute(false, false, true),
			"skip_deprovision_on_destroy": common.GetSkipProvisioningOnDestroyAttribute(false, false, true),
			"timeouts":                    common.GetTimeoutsAttribute(ctx),
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

func getAzureTenantIDAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.AZURE_TENANT_ID_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("tenant_id"),
		},
	}
}

func getCustomDomainAttribute(required, optional, computed bool) rschema.StringAttribute {
	return common.GetCustomDomainAttribute(required, optional, computed, common.AZURE_CUSTOM_DOMAIN_DESCRIPTION)
}

func getAzureSubscriptionIDAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.AZURE_SUBSCRIPTION_ID_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("subscription_id"),
		},
	}
}

func getTagsAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return common.GetTagsAttribute(required, optional, computed, common.AZURE_TAGS_DESCRIPTION)
}

func getPrivateLinkServiceAttribute(required, optional, computed bool) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.AZURE_PRIVATE_LINK_SERVICE_DESCRIPTION,
		Default:             objectdefault.StaticValue(privateLinkServiceDefaultObject),
		Attributes: map[string]rschema.Attribute{
			"allowed_subscriptions": rschema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: common.AZURE_PRIVATE_LINK_SERVICE_ALLOWED_SUBSCRIPTIONS_DESCRIPTION,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

var privateLinkServiceDefaultObject, _ = types.ObjectValue(
	map[string]attr.Type{
		"allowed_subscriptions": types.ListType{
			ElemType: types.StringType,
		},
	},
	map[string]attr.Value{
		"allowed_subscriptions": types.ListNull(types.StringType),
	},
)

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
