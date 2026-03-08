package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureEnvResourceModel struct {
	Id                    types.String                    `tfsdk:"id"`
	Name                  types.String                    `tfsdk:"name"`
	CustomDomain          types.String                    `tfsdk:"custom_domain"`
	NodeGroups            []common.NodeGroupsModel        `tfsdk:"node_groups"`
	Region                types.String                    `tfsdk:"region"`
	CIDR                  types.String                    `tfsdk:"cidr"`
	TenantID              types.String                    `tfsdk:"tenant_id"`
	SubscriptionID        types.String                    `tfsdk:"subscription_id"`
	Zones                 types.List                      `tfsdk:"zones"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	Tags                  []common.KeyValueModel          `tfsdk:"tags"`
	PrivateLinkService    *PrivateLinkServiceModel        `tfsdk:"private_link_service"`
	MetricsEndpoint       *MetricsEndpointModel           `tfsdk:"metrics_endpoint"`

	SpecRevision                 types.Int64 `tfsdk:"spec_revision"`
	ForceDestroy                 types.Bool  `tfsdk:"force_destroy"`
	ForceDestroyClusters         types.Bool  `tfsdk:"force_destroy_clusters"`
	SkipDeprovisionOnDestroy     types.Bool  `tfsdk:"skip_deprovision_on_destroy"`
	AllowDeleteWhileDisconnected types.Bool  `tfsdk:"allow_delete_while_disconnected"`
}

type PrivateLinkServiceModel struct {
	AllowedSubscriptions []types.String `tfsdk:"allowed_subscriptions"`
}

type LoadBalancersModel struct {
	Public   *PublicLoadBalancerModel   `tfsdk:"public"`
	Internal *InternalLoadBalancerModel `tfsdk:"internal"`
}

type PublicLoadBalancerModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

type InternalLoadBalancerModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

type MetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

func (e AzureEnvResourceModel) toSDK(ctx context.Context) (client.CreateAzureEnvInput, client.UpdateAzureEnvInput, diag.Diagnostics) {
	var zones []string
	var allDiags diag.Diagnostics
	if !e.Zones.IsUnknown() && !e.Zones.IsNull() {
		diags := e.Zones.ElementsAs(ctx, &zones, false)
		allDiags.Append(diags...)
	}

	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)
	nodeGroups, diags := nodeGroupsToSDK(ctx, e.NodeGroups)
	allDiags.Append(diags...)
	loadBalancingStrategy := (*client.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	metricsEndpoint := metricsEndpointToSDK(e.MetricsEndpoint)
	cloudConnect := false

	var tags []*client.KeyValueInput
	for _, t := range e.Tags {
		tags = append(tags, &client.KeyValueInput{
			Key:   t.Key.ValueString(),
			Value: t.Value.ValueString(),
		})
	}

	var allowedSubscriptions = []string{}
	if e.PrivateLinkService != nil {
		for _, as := range e.PrivateLinkService.AllowedSubscriptions {
			allowedSubscriptions = append(allowedSubscriptions, as.ValueString())
		}
	}

	create := client.CreateAzureEnvInput{
		Name: e.Name.ValueString(),
		Spec: &client.CreateAzureEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			TenantID:              e.TenantID.ValueString(),
			SubscriptionID:        e.SubscriptionID.ValueString(),
			Region:                e.Region.ValueString(),
			Cidr:                  e.CIDR.ValueString(),
			Zones:                 zones,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			CloudConnect:          &cloudConnect,
			Tags:                  tags,
			PrivateLinkService: &client.PrivateLinkServiceSpecInput{
				AllowedSubscriptions: allowedSubscriptions,
			},
			MetricsEndpoint: metricsEndpoint,
		},
	}

	strategy := client.UpdateStrategyReplace
	update := client.UpdateAzureEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &client.UpdateAzureEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			Zones:                 zones,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			Tags:                  tags,
			PrivateLinkService: &client.PrivateLinkServiceSpecInput{
				AllowedSubscriptions: allowedSubscriptions,
			},
			MetricsEndpoint: metricsEndpoint,
		},
	}

	return create, update, allDiags
}

func (model *AzureEnvResourceModel) toModel(env client.GetAzureEnv_AzureEnv) diag.Diagnostics {
	var allDiags diag.Diagnostics
	model.Name = types.StringValue(env.Name)
	model.Region = types.StringValue(env.Spec.Region)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.SubscriptionID = types.StringValue(env.Spec.SubscriptionID)
	model.TenantID = types.StringValue(env.Spec.TenantID)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)

	nodeGroups, diags := nodeGroupsToModel(env.Spec.NodeGroups)
	allDiags.Append(diags...)
	model.NodeGroups = nodeGroups

	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)

	zones, diags := common.ListToModel(env.Spec.Zones)
	allDiags.Append(diags...)
	model.Zones = zones

	model.MetricsEndpoint = metricsEndpointToModel(&env.Spec.MetricsEndpoint)

	var tags []common.KeyValueModel
	for _, t := range env.Spec.Tags {
		tags = append(tags, common.KeyValueModel{
			Key:   types.StringValue(t.Key),
			Value: types.StringValue(t.Value),
		})
	}

	model.PrivateLinkService = &PrivateLinkServiceModel{
		AllowedSubscriptions: common.ListStringToModel(env.Spec.PrivateLinkService.AllowedSubscriptions),
	}
	model.SpecRevision = types.Int64Value(env.SpecRevision)
	model.Tags = tags
	return allDiags
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *client.AzureEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *client.AzureEnvLoadBalancerPublicSpecInput
	var internal *client.AzureEnvLoadBalancerInternalSpecInput

	if loadBalancers.Public != nil {
		public = &client.AzureEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
		}
	}

	if loadBalancers.Internal != nil {
		internal = &client.AzureEnvLoadBalancerInternalSpecInput{
			Enabled:        loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
		}
	}

	return &client.AzureEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers client.AzureEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
	model := &LoadBalancersModel{}

	var publicSourceIpRanges []types.String
	for _, s := range loadBalancers.Public.SourceIPRanges {
		publicSourceIpRanges = append(publicSourceIpRanges, types.StringValue(s))
	}

	model.Public = &PublicLoadBalancerModel{
		Enabled:        types.BoolValue(loadBalancers.Public.Enabled),
		SourceIPRanges: publicSourceIpRanges,
	}

	var internalSourceIpRanges []types.String
	for _, s := range loadBalancers.Internal.SourceIPRanges {
		internalSourceIpRanges = append(internalSourceIpRanges, types.StringValue(s))
	}

	model.Internal = &InternalLoadBalancerModel{
		Enabled:        types.BoolValue(loadBalancers.Internal.Enabled),
		SourceIPRanges: internalSourceIpRanges,
	}

	return model
}

func nodeGroupsToSDK(ctx context.Context, nodeGroups []common.NodeGroupsModel) ([]*client.AzureEnvNodeGroupSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var sdkNodeGroups []*client.AzureEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []client.NodeReservation
		diags := np.Reservations.ElementsAs(ctx, &reservations, false)
		allDiags.Append(diags...)

		var zones []string
		if !np.Zones.IsUnknown() && !np.Zones.IsNull() {
			diags = np.Zones.ElementsAs(ctx, &zones, false)
			allDiags.Append(diags...)
		}

		sdkNodeGroups = append(sdkNodeGroups, &client.AzureEnvNodeGroupSpecInput{
			Name:            np.Name.ValueStringPointer(),
			NodeType:        np.NodeType.ValueString(),
			Zones:           zones,
			Reservations:    reservations,
			CapacityPerZone: np.CapacityPerZone.ValueInt64(),
		})
	}

	return sdkNodeGroups, allDiags
}

func nodeGroupsToModel(nodeGroups []*client.AzureEnvSpecFragment_NodeGroups) ([]common.NodeGroupsModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var modelNodeGroups []common.NodeGroupsModel
	for _, np := range nodeGroups {
		zones, diags := common.ListToModel(np.Zones)
		allDiags.Append(diags...)
		reservations, diags := common.ReservationsToModel(np.Reservations)
		allDiags.Append(diags...)

		modelNodeGroups = append(modelNodeGroups, common.NodeGroupsModel{
			Name:            types.StringValue(np.Name),
			NodeType:        types.StringValue(np.NodeType),
			Zones:           zones,
			Reservations:    reservations,
			CapacityPerZone: types.Int64Value(np.CapacityPerZone),
		})
	}

	return modelNodeGroups, allDiags
}

func maintenanceWindowsToModel(input []*client.AzureEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
	var maintenanceWindow []common.MaintenanceWindowModel
	for _, mw := range input {
		var days []types.String
		for _, day := range mw.Days {
			days = append(days, types.StringValue(string(day)))
		}

		maintenanceWindow = append(maintenanceWindow, common.MaintenanceWindowModel{
			Name:          types.StringValue(mw.Name),
			Enabled:       types.BoolValue(mw.Enabled),
			Hour:          types.Int64Value(mw.Hour),
			LengthInHours: types.Int64Value(mw.LengthInHours),
			Days:          days,
		})
	}

	return maintenanceWindow
}

func metricsEndpointToSDK(endpoint *MetricsEndpointModel) *client.MetricsEndpointSpecInput {
	if endpoint == nil {
		return nil
	}

	var sourceIPRanges []string
	for _, ip := range endpoint.SourceIPRanges {
		sourceIPRanges = append(sourceIPRanges, ip.ValueString())
	}

	return &client.MetricsEndpointSpecInput{
		Enabled:        endpoint.Enabled.ValueBoolPointer(),
		SourceIPRanges: sourceIPRanges,
	}
}

func metricsEndpointToModel(endpoint *client.AzureEnvSpecFragment_MetricsEndpoint) *MetricsEndpointModel {
	if endpoint == nil {
		return nil
	}

	var sourceIPRanges []types.String
	for _, ip := range endpoint.SourceIPRanges {
		sourceIPRanges = append(sourceIPRanges, types.StringValue(ip))
	}

	return &MetricsEndpointModel{
		Enabled:        types.BoolValue(endpoint.Enabled),
		SourceIPRanges: sourceIPRanges,
	}
}
