package env

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/modifiers"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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

func (r *HCloudEnvResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) HCloud environment resource.`),
		Attributes: map[string]rschema.Attribute{
			"id":                              common.IDAttribute,
			"name":                            common.NameAttribute,
			"hcloud_token_enc":                getHCloudTokenEncAttribute(true, false, false),
			"custom_domain":                   common.GetCommonCustomDomainAttribute(false, true, false),
			"load_balancers":                  getLoadBalancersAttribute(false, true, true),
			"load_balancing_strategy":         common.GetLoadBalancingStrategyAttribute(false, true, true),
			"maintenance_windows":             common.GetMaintenanceWindowAttribute(false, true, false),
			"cidr":                            common.GetCIDRAttribute(true, false, false),
			"locations":                       common.GetZonesAttribute(false, true, true, common.HCLOUD_LOCATIONS_DESCRIPTION),
			"node_groups":                     getNodeGroupsAttribute(true, false, false),
			"network_zone":                    common.GetRegionAttribute(true, false, false, common.HCLOUD_NETWORK_ZONE_DESCRIPTION),
			"wireguard_peers":                 getWireguardPeersAttribute(false, true, false),
			"spec_revision":                   common.SpecRevisionAttribute,
			"force_destroy":                   common.GetForceDestroyAttribute(false, true, true),
			"force_destroy_clusters":          common.GetForceDestroyClustersAttribute(false, true, true),
			"skip_deprovision_on_destroy":     common.GetSkipProvisioningOnDestroyAttribute(false, true, true),
			"allow_delete_while_disconnected": common.GetAllowDeleteWhileDisconnectedAttribute(false, true, true),
		},
	}
}

func (d *HCloudEnvDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: heredoc.Doc(`Bring Your Own Cloud (BYOC) HCloud environment data source.`),
		Attributes: map[string]dschema.Attribute{
			"id":                      common.IDAttribute,
			"name":                    common.NameAttribute,
			"hcloud_token_enc":        getHCloudTokenEncAttribute(false, false, true),
			"custom_domain":           common.GetCommonCustomDomainAttribute(false, false, true),
			"load_balancers":          getLoadBalancersAttribute(false, false, true),
			"load_balancing_strategy": common.GetLoadBalancingStrategyAttribute(false, false, true),
			"maintenance_windows":     common.GetMaintenanceWindowAttribute(false, false, true),
			"cidr":                    common.GetCIDRAttribute(false, false, true),
			"locations":               common.GetZonesAttribute(false, false, true, common.HCLOUD_LOCATIONS_DESCRIPTION),
			"node_groups":             getNodeGroupsAttribute(false, false, true),
			"network_zone":            common.GetRegionAttribute(false, false, true, common.HCLOUD_NETWORK_ZONE_DESCRIPTION),
			"spec_revision":           common.SpecRevisionAttribute,
			"wireguard_peers":         getWireguardPeersAttribute(false, false, true),

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

func getHCloudTokenEncAttribute(required, optional, computed bool) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.HCLOUD_TOKEN_ENC_DESCRIPTION,
		PlanModifiers: []planmodifier.String{
			modifiers.ImmutableString("hcloud_token_enc"),
		},
	}
}

func getNodeGroupsAttribute(required, optional, computed bool) rschema.SetNestedAttribute {
	return rschema.SetNestedAttribute{
		NestedObject:        nodeGroupAttribute,
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.NODE_GROUP_DESCRIPTION,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
	}
}

func getWireguardPeersAttribute(required, optional, computed bool) rschema.ListNestedAttribute {
	return rschema.ListNestedAttribute{
		NestedObject:        wireguardPeersAttribute,
		Optional:            optional,
		Required:            required,
		Computed:            computed,
		MarkdownDescription: common.HCLOUD_WIREGUARD_PEERS_DESCRIPTION,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
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
			MarkdownDescription: common.NODE_GROUP_DESCRIPTION,
		},
		"capacity_per_location": rschema.Int64Attribute{
			Required:            true,
			MarkdownDescription: common.NODE_GROUP_CAPACITY_PER_ZONE_DESCRIPTION,
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"locations": rschema.ListAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: common.NODE_GROUP_ZONES_DESCRIPTION,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"reservations": common.GetReservationsAttribute(true, false, false),
	},
}

var wireguardPeersAttribute = rschema.NestedAttributeObject{
	Attributes: map[string]rschema.Attribute{
		"public_key": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.HCLOUD_WIREGUARD_PEERS_PUBLIC_KEY_DESCRIPTION,
		},
		"allowed_ips": rschema.ListAttribute{
			ElementType:         types.StringType,
			Required:            true,
			MarkdownDescription: common.HCLOUD_WIREGUARD_PEERS_ALLOWED_IPS_DESCRIPTION,
		},
		"endpoint": rschema.StringAttribute{
			Required:            true,
			MarkdownDescription: common.HCLOUD_WIREGUARD_PEERS_ENDPOINT_DESCRIPTION,
		},
	},
}
