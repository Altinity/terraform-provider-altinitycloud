package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HCloudEnvResourceModel struct {
	Id                    types.String                    `tfsdk:"id"`
	Name                  types.String                    `tfsdk:"name"`
	HCloudTokenEnc        types.String                    `tfsdk:"hcloud_token_enc"`
	CustomDomain          types.String                    `tfsdk:"custom_domain"`
	NodeGroups            []NodeGroupsModel               `tfsdk:"node_groups"`
	NetworkZone           types.String                    `tfsdk:"network_zone"`
	CIDR                  types.String                    `tfsdk:"cidr"`
	Locations             types.List                      `tfsdk:"locations"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	WireguardPeers        []WireguardPeers                `tfsdk:"wireguard_peers"`
	MetricsEndpoint       *MetricsEndpointModel           `tfsdk:"metrics_endpoint"`

	SpecRevision                 types.Int64    `tfsdk:"spec_revision"`
	ForceDestroy                 types.Bool     `tfsdk:"force_destroy"`
	ForceDestroyClusters         types.Bool     `tfsdk:"force_destroy_clusters"`
	SkipDeprovisionOnDestroy     types.Bool     `tfsdk:"skip_deprovision_on_destroy"`
	AllowDeleteWhileDisconnected types.Bool     `tfsdk:"allow_delete_while_disconnected"`
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
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

type NodeGroupsModel struct {
	Name                types.String `tfsdk:"name"`
	NodeType            types.String `tfsdk:"node_type"`
	CapacityPerLocation types.Int64  `tfsdk:"capacity_per_location"`
	Locations           types.List   `tfsdk:"locations"`
	Reservations        types.Set    `tfsdk:"reservations"`
}

type WireguardPeers struct {
	PublicKey  types.String `tfsdk:"public_key"`
	AllowedIPs types.List   `tfsdk:"allowed_ips"`
	Endpoint   types.String `tfsdk:"endpoint"`
}

type MetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

func (e HCloudEnvResourceModel) toSDK(ctx context.Context) (client.CreateHCloudEnvInput, client.UpdateHCloudEnvInput, diag.Diagnostics) {
	var locations []string
	var allDiags diag.Diagnostics
	if !e.Locations.IsUnknown() && !e.Locations.IsNull() {
		diags := e.Locations.ElementsAs(ctx, &locations, false)
		allDiags.Append(diags...)
	}

	wireguardPeers, diags := wireguardPeersToSDK(ctx, e.WireguardPeers)
	allDiags.Append(diags...)

	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)

	nodeGroups, diags := nodeGroupsToSDK(ctx, e.NodeGroups)
	allDiags.Append(diags...)

	loadBalancingStrategy := (*client.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	metricsEndpoint := metricsEndpointToSDK(e.MetricsEndpoint)
	cloudConnect := false

	create := client.CreateHCloudEnvInput{
		Name: e.Name.ValueString(),
		Spec: &client.CreateHCloudEnvSpecInput{
			HcloudTokenEnc:        e.HCloudTokenEnc.ValueString(),
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			NetworkZone:           e.NetworkZone.ValueString(),
			Cidr:                  e.CIDR.ValueString(),
			Locations:             locations,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			CloudConnect:          &cloudConnect,
			WireguardPeers:        wireguardPeers,
			MetricsEndpoint:       metricsEndpoint,
		},
	}

	strategy := client.UpdateStrategyReplace
	update := client.UpdateHCloudEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &client.UpdateHCloudEnvSpecInput{
			HcloudTokenEnc:        e.HCloudTokenEnc.ValueStringPointer(),
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			WireguardPeers:        wireguardPeers,
			MetricsEndpoint:       metricsEndpoint,
		},
	}

	return create, update, allDiags
}

func (model *HCloudEnvResourceModel) toModel(env client.GetHCloudEnv_HcloudEnv) diag.Diagnostics {
	var allDiags diag.Diagnostics

	model.Name = types.StringValue(env.Name)
	model.NetworkZone = types.StringValue(env.Spec.NetworkZone)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)

	nodeGroups, diags := nodeGroupsToModel(env.Spec.NodeGroups)
	allDiags.Append(diags...)
	model.NodeGroups = nodeGroups

	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)

	locations, diags := common.ListToModel(env.Spec.Locations)
	allDiags.Append(diags...)
	model.Locations = locations

	wireguardPeers, diags := wireguardPeersToModel(env.Spec.WireguardPeers)
	allDiags.Append(diags...)
	model.WireguardPeers = wireguardPeers

	model.MetricsEndpoint = metricsEndpointToModel(&env.Spec.MetricsEndpoint)
	model.SpecRevision = types.Int64Value(env.SpecRevision)

	return allDiags
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *client.HCloudEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *client.HCloudEnvLoadBalancerPublicSpecInput
	var internal *client.HCloudEnvLoadBalancerInternalSpecInput

	if loadBalancers.Public != nil {
		public = &client.HCloudEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
		}
	}

	if loadBalancers.Internal != nil {
		internal = &client.HCloudEnvLoadBalancerInternalSpecInput{
			Enabled:        loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
		}
	}

	return &client.HCloudEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers client.HCloudEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
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

func nodeGroupsToSDK(ctx context.Context, nodeGroups []NodeGroupsModel) ([]*client.HCloudEnvNodeGroupSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var sdkNodeGroups []*client.HCloudEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []client.NodeReservation
		if !np.Reservations.IsUnknown() && !np.Reservations.IsNull() {
			diags := np.Reservations.ElementsAs(ctx, &reservations, false)
			allDiags.Append(diags...)
		}

		var locations []string
		if !np.Locations.IsUnknown() && !np.Locations.IsNull() {
			diags := np.Locations.ElementsAs(ctx, &locations, false)
			allDiags.Append(diags...)
		}

		sdkNodeGroups = append(sdkNodeGroups, &client.HCloudEnvNodeGroupSpecInput{
			Name:                np.Name.ValueStringPointer(),
			NodeType:            np.NodeType.ValueString(),
			Locations:           locations,
			Reservations:        reservations,
			CapacityPerLocation: np.CapacityPerLocation.ValueInt64(),
		})
	}

	return sdkNodeGroups, allDiags
}

func wireguardPeersToSDK(ctx context.Context, peers []WireguardPeers) ([]*client.HCloudEnvWireguardPeerSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var sdkPeers []*client.HCloudEnvWireguardPeerSpecInput
	for _, p := range peers {
		var allowedIPs []string
		if !p.AllowedIPs.IsUnknown() && !p.AllowedIPs.IsNull() {
			diags := p.AllowedIPs.ElementsAs(ctx, &allowedIPs, false)
			allDiags.Append(diags...)
		}

		sdkPeers = append(sdkPeers, &client.HCloudEnvWireguardPeerSpecInput{
			PublicKey:  p.PublicKey.ValueString(),
			AllowedIPs: allowedIPs,
			Endpoint:   p.Endpoint.ValueString(),
		})
	}

	return sdkPeers, allDiags
}

func nodeGroupsToModel(nodeGroups []*client.HCloudEnvSpecFragment_NodeGroups) ([]NodeGroupsModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var modelNodeGroups []NodeGroupsModel
	for _, np := range nodeGroups {
		locations, diags := common.ListToModel(np.Locations)
		allDiags.Append(diags...)
		reservations, diags := common.ReservationsToModel(np.Reservations)
		allDiags.Append(diags...)

		modelNodeGroups = append(modelNodeGroups, NodeGroupsModel{
			Name:                types.StringValue(np.Name),
			NodeType:            types.StringValue(np.NodeType),
			Locations:           locations,
			Reservations:        reservations,
			CapacityPerLocation: types.Int64Value(np.CapacityPerLocation),
		})
	}

	return modelNodeGroups, allDiags
}

func maintenanceWindowsToModel(input []*client.HCloudEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
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

func wireguardPeersToModel(input []*client.HCloudEnvSpecFragment_WireguardPeers) ([]WireguardPeers, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var peers []WireguardPeers
	for _, p := range input {
		allowedIPs, diags := common.ListToModel(p.AllowedIPs)
		allDiags.Append(diags...)

		peers = append(peers, WireguardPeers{
			PublicKey:  types.StringValue(p.PublicKey),
			AllowedIPs: allowedIPs,
			Endpoint:   types.StringValue(p.Endpoint),
		})
	}

	return peers, allDiags
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

func metricsEndpointToModel(endpoint *client.HCloudEnvSpecFragment_MetricsEndpoint) *MetricsEndpointModel {
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
