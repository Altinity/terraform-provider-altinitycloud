package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPEnvResourceModel struct {
	Id           types.String             `tfsdk:"id"`
	Name         types.String             `tfsdk:"name"`
	CustomDomain types.String             `tfsdk:"custom_domain"`
	NodeGroups   []common.NodeGroupsModel `tfsdk:"node_groups"`

	Region                types.String                    `tfsdk:"region"`
	CIDR                  types.String                    `tfsdk:"cidr"`
	GCPProjectID          types.String                    `tfsdk:"gcp_project_id"`
	Zones                 types.List                      `tfsdk:"zones"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	PeeringConnections    []GCPEnvPeeringConnectionModel  `tfsdk:"peering_connections"`

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

type GCPEnvPeeringConnectionModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	NetworkName types.String `tfsdk:"network_name"`
}

func (e GCPEnvResourceModel) toSDK() (sdk.CreateGCPEnvInput, sdk.UpdateGCPEnvInput) {
	var zones []string
	e.Zones.ElementsAs(context.TODO(), &zones, false)

	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)
	nodeGroups := nodeGroupsToSDK(e.NodeGroups)
	loadBalancingStrategy := (*sdk.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	cloudConnect := false

	var peeringConnections []*sdk.GCPEnvPeeringConnectionSpecInput
	for _, p := range e.PeeringConnections {
		peeringConnections = append(peeringConnections, &sdk.GCPEnvPeeringConnectionSpecInput{
			ProjectID:   p.ProjectID.ValueStringPointer(),
			NetworkName: p.NetworkName.ValueString(),
		})
	}

	create := sdk.CreateGCPEnvInput{
		Name: e.Name.ValueString(),
		Spec: &sdk.CreateGCPEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			GCPProjectID:          e.GCPProjectID.ValueString(),
			Region:                e.Region.ValueString(),
			Cidr:                  e.CIDR.ValueString(),
			Zones:                 zones,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			CloudConnect:          &cloudConnect,
			PeeringConnections:    peeringConnections,
		},
	}

	strategy := sdk.UpdateStrategyReplace
	update := sdk.UpdateGCPEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &sdk.UpdateGCPEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			NodeGroups:            nodeGroups,
			Zones:                 zones,
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			MaintenanceWindows:    maintenanceWindows,
			PeeringConnections:    peeringConnections,
		},
	}

	return create, update
}

func (model *GCPEnvResourceModel) toModel(env sdk.GetGCPEnv_GCPEnv) {
	model.Name = types.StringValue(env.Name)
	model.Region = types.StringValue(env.Spec.Region)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.GCPProjectID = types.StringValue(env.Spec.GCPProjectID)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)
	model.NodeGroups = nodeGroupsToModel(env.Spec.NodeGroups)
	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)
	model.Zones = common.ListToModel(env.Spec.Zones)

	var peeringConnections []GCPEnvPeeringConnectionModel
	for _, p := range env.Spec.PeeringConnections {
		peeringConnections = append(peeringConnections, GCPEnvPeeringConnectionModel{
			ProjectID:   types.StringPointerValue(p.ProjectID),
			NetworkName: types.StringValue(p.NetworkName),
		})
	}
	model.PeeringConnections = peeringConnections
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *sdk.GCPEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *sdk.GCPEnvLoadBalancerPublicSpecInput
	var internal *sdk.GCPEnvLoadBalancerInternalSpecInput

	if loadBalancers.Public != nil {
		public = &sdk.GCPEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
		}
	}

	if loadBalancers.Internal != nil {
		internal = &sdk.GCPEnvLoadBalancerInternalSpecInput{
			Enabled:        loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
		}
	}

	return &sdk.GCPEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers sdk.GCPEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
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

func nodeGroupsToSDK(nodeGroups []common.NodeGroupsModel) []*sdk.GCPEnvNodeGroupSpecInput {
	var sdkNodeGroups []*sdk.GCPEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []sdk.NodeReservation
		np.Reservations.ElementsAs(context.TODO(), &reservations, false)

		var zones []string
		np.Zones.ElementsAs(context.TODO(), &zones, false)

		sdkNodeGroups = append(sdkNodeGroups, &sdk.GCPEnvNodeGroupSpecInput{
			Name:            np.Name.ValueStringPointer(),
			NodeType:        np.NodeType.ValueString(),
			Zones:           zones,
			Reservations:    reservations,
			CapacityPerZone: np.CapacityPerZone.ValueInt64(),
		})
	}

	return sdkNodeGroups
}

func nodeGroupsToModel(nodeGroups []*sdk.GCPEnvSpecFragment_NodeGroups) []common.NodeGroupsModel {
	var modelNodeGroups []common.NodeGroupsModel
	for _, np := range nodeGroups {
		modelNodeGroups = append(modelNodeGroups, common.NodeGroupsModel{
			Name:            types.StringValue(np.Name),
			NodeType:        types.StringValue(np.NodeType),
			Zones:           common.ListToModel(np.Zones),
			Reservations:    common.ReservationsToModel(np.Reservations),
			CapacityPerZone: types.Int64Value(np.CapacityPerZone),
		})
	}

	return modelNodeGroups
}

func maintenanceWindowsToModel(input []*sdk.GCPEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
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

func reorderNodeGroups(model []common.NodeGroupsModel, input []*sdk.GCPEnvSpecFragment_NodeGroups) []*sdk.GCPEnvSpecFragment_NodeGroups {
	orderedNodeGroups := make([]*sdk.GCPEnvSpecFragment_NodeGroups, 0, len(input))

	for _, ng := range model {
		for _, apiGroup := range input {
			if ng.NodeType.ValueString() == apiGroup.NodeType {
				orderedNodeGroups = append(orderedNodeGroups, apiGroup)
				break
			}
		}
	}

	return orderedNodeGroups
}
