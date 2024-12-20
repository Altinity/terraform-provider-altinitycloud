package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEnvResourceModel struct {
	Id                    types.String                    `tfsdk:"id"`
	Name                  types.String                    `tfsdk:"name"`
	CustomDomain          types.String                    `tfsdk:"custom_domain"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	Region                types.String                    `tfsdk:"region"`
	NAT                   types.Bool                      `tfsdk:"nat"`
	NumberOfZones         types.Int64                     `tfsdk:"number_of_zones"`
	CIDR                  types.String                    `tfsdk:"cidr"`
	AWSAccountID          types.String                    `tfsdk:"aws_account_id"`
	Zones                 types.List                      `tfsdk:"zones"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	NodeGroups            []common.NodeGroupsModel        `tfsdk:"node_groups"`
	PeeringConnections    []AWSEnvPeeringConnectionModel  `tfsdk:"peering_connections"`
	Endpoints             []AWSEnvEndpointModel           `tfsdk:"endpoints"`
	Tags                  []common.KeyValueModel          `tfsdk:"tags"`
	CloudConnect          types.Bool                      `tfsdk:"cloud_connect"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`

	SpecRevision             types.Int64 `tfsdk:"spec_revision"`
	ForceDestroy             types.Bool  `tfsdk:"force_destroy"`
	ForceDestroyClusters     types.Bool  `tfsdk:"force_destroy_clusters"`
	SkipDeprovisionOnDestroy types.Bool  `tfsdk:"skip_deprovision_on_destroy"`
}

type LoadBalancersModel struct {
	Public   *PublicLoadBalancerModel   `tfsdk:"public"`
	Internal *InternalLoadBalancerModel `tfsdk:"internal"`
}

type PublicLoadBalancerModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
	CrossZone      types.Bool     `tfsdk:"cross_zone"`
}

type InternalLoadBalancerModel struct {
	Enabled                          types.Bool     `tfsdk:"enabled"`
	SourceIPRanges                   []types.String `tfsdk:"source_ip_ranges"`
	CrossZone                        types.Bool     `tfsdk:"cross_zone"`
	EndpointServiceAllowedPrincipals []types.String `tfsdk:"endpoint_service_allowed_principals"`
}

type AWSEnvEndpointModel struct {
	ServiceName types.String `tfsdk:"service_name"`
	Alias       types.String `tfsdk:"alias"`
	PrivateDNS  types.Bool   `tfsdk:"private_dns"`
}

type AWSEnvPeeringConnectionModel struct {
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	VpcID        types.String `tfsdk:"vpc_id"`
	VpcRegion    types.String `tfsdk:"vpc_region"`
}

func (e AWSEnvResourceModel) toSDK() (sdk.CreateAWSEnvInput, sdk.UpdateAWSEnvInput) {
	var zones []string
	e.Zones.ElementsAs(context.TODO(), &zones, false)

	var peeringConnections []*sdk.AWSEnvPeeringConnectionSpecInput
	for _, p := range e.PeeringConnections {
		peeringConnections = append(peeringConnections, &sdk.AWSEnvPeeringConnectionSpecInput{
			AwsAccountID: p.AWSAccountID.ValueStringPointer(),
			VpcID:        p.VpcID.ValueString(),
			VpcRegion:    p.VpcRegion.ValueStringPointer(),
		})
	}

	var endpoints []*sdk.AWSEnvEndpointSpecInput
	for _, e := range e.Endpoints {
		endpoints = append(endpoints, &sdk.AWSEnvEndpointSpecInput{
			ServiceName: e.ServiceName.ValueString(),
			Alias:       e.Alias.ValueStringPointer(),
		})
	}

	var tags []*sdk.KeyValueInput
	for _, t := range e.Tags {
		tags = append(tags, &sdk.KeyValueInput{
			Key:   t.Key.ValueString(),
			Value: t.Value.ValueString(),
		})
	}

	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)
	nodeGroups := nodeGroupsToSDK(e.NodeGroups)
	loadBalancingStrategy := (*sdk.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	cloudConnect := e.CloudConnect.ValueBool()

	create := sdk.CreateAWSEnvInput{
		Name: e.Name.ValueString(),
		Spec: &sdk.CreateAWSEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			NodeGroups:            nodeGroups,
			Region:                e.Region.ValueString(),
			Nat:                   e.NAT.ValueBoolPointer(),
			NumberOfZones:         e.NumberOfZones.ValueInt64Pointer(),
			AwsAccountID:          e.AWSAccountID.ValueString(),
			Cidr:                  e.CIDR.ValueString(),
			Zones:                 zones,
			PeeringConnections:    peeringConnections,
			Endpoints:             endpoints,
			Tags:                  tags,
			CloudConnect:          &cloudConnect,
			MaintenanceWindows:    maintenanceWindows,
		},
	}

	strategy := client.UpdateStrategyReplace
	update := sdk.UpdateAWSEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &sdk.AWSEnvUpdateSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			NodeGroups:            nodeGroups,
			NumberOfZones:         e.NumberOfZones.ValueInt64Pointer(),
			Zones:                 zones,
			PeeringConnections:    peeringConnections,
			Endpoints:             endpoints,
			Tags:                  tags,
			MaintenanceWindows:    maintenanceWindows,
		},
	}

	return create, update
}

func (model *AWSEnvResourceModel) toModel(env sdk.GetAWSEnv_AwsEnv) {
	model.Name = types.StringValue(env.Name)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.Region = types.StringValue(env.Spec.Region)
	model.NAT = types.BoolValue(env.Spec.Nat)
	model.AWSAccountID = types.StringValue(env.Spec.AwsAccountID)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)
	model.NodeGroups = nodeGroupsToModel(env.Spec.NodeGroups)
	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)
	model.Zones = common.ListToModel(env.Spec.Zones)

	var peeringConnections []AWSEnvPeeringConnectionModel
	for _, p := range env.Spec.PeeringConnections {
		peeringConnections = append(peeringConnections, AWSEnvPeeringConnectionModel{
			AWSAccountID: types.StringPointerValue(p.AwsAccountID),
			VpcID:        types.StringValue(p.VpcID),
			VpcRegion:    types.StringPointerValue(p.VpcRegion),
		})
	}
	model.PeeringConnections = peeringConnections

	var endpoints []AWSEnvEndpointModel
	for _, e := range env.Spec.Endpoints {
		endpoints = append(endpoints, AWSEnvEndpointModel{
			ServiceName: types.StringValue(e.ServiceName),
			Alias:       types.StringPointerValue(e.Alias),
			PrivateDNS:  types.BoolValue(e.PrivateDNS),
		})
	}
	model.Endpoints = endpoints

	var tags []common.KeyValueModel
	for _, t := range env.Spec.Tags {
		tags = append(tags, common.KeyValueModel{
			Key:   types.StringValue(t.Key),
			Value: types.StringValue(t.Value),
		})
	}

	model.Tags = tags

	model.CloudConnect = types.BoolValue(env.Spec.CloudConnect)
	model.SpecRevision = types.Int64Value(env.SpecRevision)
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *sdk.AWSEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *sdk.AWSEnvLoadBalancerPublicSpecInput
	var internal *sdk.AWSEnvLoadBalancerInternalSpecInput

	if loadBalancers.Public != nil {
		public = &sdk.AWSEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
			CrossZone:      loadBalancers.Public.CrossZone.ValueBoolPointer(),
		}
	}

	if loadBalancers.Internal != nil {
		var endpointServiceAllowedPrincipals []string
		for _, ap := range loadBalancers.Internal.EndpointServiceAllowedPrincipals {
			endpointServiceAllowedPrincipals = append(endpointServiceAllowedPrincipals, ap.ValueString())
		}

		internal = &sdk.AWSEnvLoadBalancerInternalSpecInput{
			Enabled:                          loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges:                   common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
			CrossZone:                        loadBalancers.Internal.CrossZone.ValueBoolPointer(),
			EndpointServiceAllowedPrincipals: endpointServiceAllowedPrincipals,
		}
	}

	return &sdk.AWSEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers sdk.AWSEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
	model := &LoadBalancersModel{}

	var publicSourceIpRanges []types.String
	for _, s := range loadBalancers.Public.SourceIPRanges {
		publicSourceIpRanges = append(publicSourceIpRanges, types.StringValue(s))
	}

	model.Public = &PublicLoadBalancerModel{
		Enabled:        types.BoolValue(loadBalancers.Public.Enabled),
		SourceIPRanges: publicSourceIpRanges,
		CrossZone:      types.BoolValue(loadBalancers.Public.CrossZone),
	}

	var internalSourceIpRanges []types.String
	for _, s := range loadBalancers.Internal.SourceIPRanges {
		internalSourceIpRanges = append(internalSourceIpRanges, types.StringValue(s))
	}

	var endpointServiceAllowedPrincipals []types.String
	for _, e := range loadBalancers.Internal.EndpointServiceAllowedPrincipals {
		endpointServiceAllowedPrincipals = append(endpointServiceAllowedPrincipals, types.StringValue(e))
	}

	model.Internal = &InternalLoadBalancerModel{
		Enabled:                          types.BoolValue(loadBalancers.Internal.Enabled),
		SourceIPRanges:                   internalSourceIpRanges,
		CrossZone:                        types.BoolValue(loadBalancers.Internal.CrossZone),
		EndpointServiceAllowedPrincipals: endpointServiceAllowedPrincipals,
	}

	return model
}

func nodeGroupsToSDK(nodeGroups []common.NodeGroupsModel) []*sdk.AWSEnvNodeGroupSpecInput {
	var sdkNodeGroups []*sdk.AWSEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []client.NodeReservation
		np.Reservations.ElementsAs(context.TODO(), &reservations, false)

		var zones []string
		np.Zones.ElementsAs(context.TODO(), &zones, false)

		sdkNodeGroups = append(sdkNodeGroups, &sdk.AWSEnvNodeGroupSpecInput{
			Name:            np.Name.ValueStringPointer(),
			NodeType:        np.NodeType.ValueString(),
			Zones:           zones,
			Reservations:    reservations,
			CapacityPerZone: np.CapacityPerZone.ValueInt64(),
		})
	}

	return sdkNodeGroups
}

func nodeGroupsToModel(nodeGroups []*sdk.AWSEnvSpecFragment_NodeGroups) []common.NodeGroupsModel {
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

func maintenanceWindowsToModel(input []*sdk.AWSEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
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

func reorderNodeGroups(model []common.NodeGroupsModel, sdk []*client.AWSEnvSpecFragment_NodeGroups) []*client.AWSEnvSpecFragment_NodeGroups {
	orderedNodeGroups := make([]*client.AWSEnvSpecFragment_NodeGroups, 0, len(sdk))

	for _, ng := range model {
		for _, apiGroup := range sdk {
			if ng.NodeType.ValueString() == apiGroup.NodeType {
				orderedNodeGroups = append(orderedNodeGroups, apiGroup)
				break
			}
		}
	}

	return orderedNodeGroups
}
