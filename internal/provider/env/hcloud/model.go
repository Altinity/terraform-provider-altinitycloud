package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
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
	Locations             types.Set                       `tfsdk:"locations"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	WireguardPeers        []WireguardPeers                `tfsdk:"wireguard_peers"`

	SpecRevision                 types.Int64 `tfsdk:"spec_revision"`
	ForceDestroy                 types.Bool  `tfsdk:"force_destroy"`
	ForceDestroyClusters         types.Bool  `tfsdk:"force_destroy_clusters"`
	SkipDeprovisionOnDestroy     types.Bool  `tfsdk:"skip_deprovision_on_destroy"`
	AllowDeleteWhileDisconnected types.Bool  `tfsdk:"allow_delete_while_disconnected"`
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
	Locations           types.Set    `tfsdk:"locations"`
	Reservations        types.Set    `tfsdk:"reservations"`
}

type WireguardPeers struct {
	publicKey  types.String `tfsdk:"public_key"`
	allowedIPs types.List   `tfsdk:"allowed_ips"`
	endpoint   types.String `tfsdk:"endpoint"`
}

func (e HCloudEnvResourceModel) toSDK() (client.CreateHCloudEnvInput, client.UpdateHCloudEnvInput) {
	var locations []string
	e.Locations.ElementsAs(context.TODO(), &locations, false)
	wireguardPeers := wireguardPeersToSDK(e.WireguardPeers)
	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)
	nodeGroups := nodeGroupsToSDK(e.NodeGroups)
	loadBalancingStrategy := (*client.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
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
		},
	}

	return create, update
}

func (model *HCloudEnvResourceModel) toModel(env client.GetHCloudEnv_HcloudEnv) {
	model.Name = types.StringValue(env.Name)
	model.NetworkZone = types.StringValue(env.Spec.NetworkZone)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)
	model.NodeGroups = nodeGroupsToModel(env.Spec.NodeGroups)
	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)
	model.Locations = common.SetToModel(env.Spec.Locations)
	model.WireguardPeers = wireguardPeersToModel(env.Spec.WireguardPeers)
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

func nodeGroupsToSDK(nodeGroups []NodeGroupsModel) []*client.HCloudEnvNodeGroupSpecInput {
	var sdkNodeGroups []*client.HCloudEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []client.NodeReservation
		np.Reservations.ElementsAs(context.TODO(), &reservations, false)

		var locations []string
		np.Locations.ElementsAs(context.TODO(), &locations, false)

		sdkNodeGroups = append(sdkNodeGroups, &client.HCloudEnvNodeGroupSpecInput{
			Name:                np.Name.ValueStringPointer(),
			NodeType:            np.NodeType.ValueString(),
			Locations:           locations,
			Reservations:        reservations,
			CapacityPerLocation: np.CapacityPerLocation.ValueInt64(),
		})
	}

	return sdkNodeGroups
}

func wireguardPeersToSDK(peers []WireguardPeers) []*client.HCloudEnvWireguardPeerSpecInput {
	var sdkPeers []*client.HCloudEnvWireguardPeerSpecInput
	for _, p := range peers {
		var allowedIPs []string
		p.allowedIPs.ElementsAs(context.TODO(), &allowedIPs, false)

		sdkPeers = append(sdkPeers, &client.HCloudEnvWireguardPeerSpecInput{
			PublicKey:  p.publicKey.ValueString(),
			AllowedIPs: allowedIPs,
			Endpoint:   p.endpoint.ValueString(),
		})
	}

	return sdkPeers
}

func nodeGroupsToModel(nodeGroups []*client.HCloudEnvSpecFragment_NodeGroups) []NodeGroupsModel {
	var modelNodeGroups []NodeGroupsModel
	for _, np := range nodeGroups {
		modelNodeGroups = append(modelNodeGroups, NodeGroupsModel{
			Name:                types.StringValue(np.Name),
			NodeType:            types.StringValue(np.NodeType),
			Locations:           common.SetToModel(np.Locations),
			Reservations:        common.ReservationsToModel(np.Reservations),
			CapacityPerLocation: types.Int64Value(np.CapacityPerLocation),
		})
	}

	return modelNodeGroups
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

func wireguardPeersToModel(input []*client.HCloudEnvSpecFragment_WireguardPeers) []WireguardPeers {
	var peers []WireguardPeers
	for _, p := range input {
		peers = append(peers, WireguardPeers{
			publicKey:  types.StringValue(p.PublicKey),
			allowedIPs: common.ListToModel(p.AllowedIPs),
			endpoint:   types.StringValue(p.Endpoint),
		})
	}

	return peers
}
