package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GCPEnvResourceModel struct {
	Id           types.String             `tfsdk:"id"`
	Name         types.String             `tfsdk:"name"`
	CustomDomain types.String             `tfsdk:"custom_domain"`
	NodeGroups   []common.NodeGroupsModel `tfsdk:"node_groups"`

	Region                  types.String                    `tfsdk:"region"`
	CIDR                    types.String                    `tfsdk:"cidr"`
	GCPProjectID            types.String                    `tfsdk:"gcp_project_id"`
	Zones                   types.List                      `tfsdk:"zones"`
	LoadBalancers           *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy   types.String                    `tfsdk:"load_balancing_strategy"`
	MaintenanceWindows      []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	PeeringConnections      []GCPEnvPeeringConnectionModel  `tfsdk:"peering_connections"`
	PrivateServiceConsumers types.List                      `tfsdk:"private_service_consumers"`
	Tags                    []common.KeyValueModel          `tfsdk:"tags"`
	MetricsEndpoint         *MetricsEndpointModel           `tfsdk:"metrics_endpoint"`

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

type GCPEnvPeeringConnectionModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	NetworkName types.String `tfsdk:"network_name"`
}

type MetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

func (e GCPEnvResourceModel) toSDK(ctx context.Context) (sdk.CreateGCPEnvInput, sdk.UpdateGCPEnvInput, diag.Diagnostics) {
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
	loadBalancingStrategy := (*sdk.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	metricsEndpoint := metricsEndpointToSDK(e.MetricsEndpoint)
	cloudConnect := false

	var tags []*sdk.KeyValueInput
	for _, t := range e.Tags {
		tags = append(tags, &sdk.KeyValueInput{
			Key:   t.Key.ValueString(),
			Value: t.Value.ValueString(),
		})
	}

	peeringConnections := make([]*sdk.GCPEnvPeeringConnectionSpecInput, 0, len(e.PeeringConnections))
	for _, p := range e.PeeringConnections {
		peeringConnections = append(peeringConnections, &sdk.GCPEnvPeeringConnectionSpecInput{
			ProjectID:   p.ProjectID.ValueStringPointer(),
			NetworkName: p.NetworkName.ValueString(),
		})
	}

	var privateServiceConsumers []string
	if !e.PrivateServiceConsumers.IsUnknown() && !e.PrivateServiceConsumers.IsNull() {
		diags := e.PrivateServiceConsumers.ElementsAs(ctx, &privateServiceConsumers, false)
		allDiags.Append(diags...)
	}

	create := sdk.CreateGCPEnvInput{
		Name: e.Name.ValueString(),
		Spec: &sdk.CreateGCPEnvSpecInput{
			CustomDomain:            e.CustomDomain.ValueStringPointer(),
			NodeGroups:              nodeGroups,
			GCPProjectID:            e.GCPProjectID.ValueString(),
			Region:                  e.Region.ValueString(),
			Cidr:                    e.CIDR.ValueString(),
			Zones:                   zones,
			LoadBalancingStrategy:   loadBalancingStrategy,
			LoadBalancers:           LoadBalancers,
			MaintenanceWindows:      maintenanceWindows,
			CloudConnect:            &cloudConnect,
			PeeringConnections:      peeringConnections,
			PrivateServiceConsumers: privateServiceConsumers,
			Tags:                    tags,
			MetricsEndpoint:         metricsEndpoint,
		},
	}

	strategy := sdk.UpdateStrategyReplace
	update := sdk.UpdateGCPEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &sdk.UpdateGCPEnvSpecInput{
			CustomDomain:            e.CustomDomain.ValueStringPointer(),
			NodeGroups:              nodeGroups,
			Zones:                   zones,
			LoadBalancingStrategy:   loadBalancingStrategy,
			LoadBalancers:           LoadBalancers,
			MaintenanceWindows:      maintenanceWindows,
			PeeringConnections:      peeringConnections,
			PrivateServiceConsumers: privateServiceConsumers,
			Tags:                    tags,
			MetricsEndpoint:         metricsEndpoint,
		},
	}

	return create, update, allDiags
}

func (model *GCPEnvResourceModel) toModel(env sdk.GetGCPEnv_GCPEnv) diag.Diagnostics {
	var allDiags diag.Diagnostics
	model.Name = types.StringValue(env.Name)
	model.Region = types.StringValue(env.Spec.Region)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.GCPProjectID = types.StringValue(env.Spec.GCPProjectID)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)

	nodeGroups, diags := nodeGroupsToModel(env.Spec.NodeGroups)
	allDiags.Append(diags...)
	model.NodeGroups = nodeGroups

	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)

	zones, diags := common.ListToModel(env.Spec.Zones)
	allDiags.Append(diags...)
	model.Zones = zones

	psc, diags := common.ListToModel(env.Spec.PrivateServiceConsumers)
	allDiags.Append(diags...)
	model.PrivateServiceConsumers = psc

	model.MetricsEndpoint = metricsEndpointToModel(&env.Spec.MetricsEndpoint)

	var tags []common.KeyValueModel
	for _, t := range env.Spec.Tags {
		tags = append(tags, common.KeyValueModel{
			Key:   types.StringValue(t.Key),
			Value: types.StringValue(t.Value),
		})
	}
	model.Tags = tags

	var peeringConnections []GCPEnvPeeringConnectionModel
	for _, p := range env.Spec.PeeringConnections {
		peeringConnections = append(peeringConnections, GCPEnvPeeringConnectionModel{
			ProjectID:   types.StringPointerValue(p.ProjectID),
			NetworkName: types.StringValue(p.NetworkName),
		})
	}
	model.PeeringConnections = peeringConnections
	model.SpecRevision = types.Int64Value(env.SpecRevision)
	return allDiags
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

func nodeGroupsToSDK(ctx context.Context, nodeGroups []common.NodeGroupsModel) ([]*sdk.GCPEnvNodeGroupSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var sdkNodeGroups []*sdk.GCPEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []sdk.NodeReservation
		if !np.Reservations.IsUnknown() && !np.Reservations.IsNull() {
			diags := np.Reservations.ElementsAs(ctx, &reservations, false)
			allDiags.Append(diags...)
		}

		var zones []string
		if !np.Zones.IsUnknown() && !np.Zones.IsNull() {
			diags := np.Zones.ElementsAs(ctx, &zones, false)
			allDiags.Append(diags...)
		}

		sdkNodeGroups = append(sdkNodeGroups, &sdk.GCPEnvNodeGroupSpecInput{
			Name:            np.Name.ValueStringPointer(),
			NodeType:        np.NodeType.ValueString(),
			Zones:           zones,
			Reservations:    reservations,
			CapacityPerZone: np.CapacityPerZone.ValueInt64(),
		})
	}

	return sdkNodeGroups, allDiags
}

func nodeGroupsToModel(nodeGroups []*sdk.GCPEnvSpecFragment_NodeGroups) ([]common.NodeGroupsModel, diag.Diagnostics) {
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

func metricsEndpointToSDK(endpoint *MetricsEndpointModel) *sdk.MetricsEndpointSpecInput {
	if endpoint == nil {
		return nil
	}

	var sourceIPRanges []string
	for _, ip := range endpoint.SourceIPRanges {
		sourceIPRanges = append(sourceIPRanges, ip.ValueString())
	}

	return &sdk.MetricsEndpointSpecInput{
		Enabled:        endpoint.Enabled.ValueBoolPointer(),
		SourceIPRanges: sourceIPRanges,
	}
}

func metricsEndpointToModel(endpoint *sdk.GCPEnvSpecFragment_MetricsEndpoint) *MetricsEndpointModel {
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
