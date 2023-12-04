package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type K8SEnvResourceModel struct {
	Id                    types.String                    `tfsdk:"id"`
	Name                  types.String                    `tfsdk:"name"`
	CustomDomain          types.String                    `tfsdk:"custom_domain"`
	LoadBalancers         *LoadBalancersModel             `tfsdk:"load_balancers"`
	LoadBalancingStrategy types.String                    `tfsdk:"load_balancing_strategy"`
	Distribution          types.String                    `tfsdk:"distribution"`
	NodeGroups            []NodeGroupsModel               `tfsdk:"node_groups"`
	CustomNodeTypes       []NodeTypeModel                 `tfsdk:"custom_node_types"`
	Logs                  *LogsModel                      `tfsdk:"logs"`
	Metrics               *MetricsModel                   `tfsdk:"metrics"`
	MaintenanceWindows    []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`

	SpecRevision             types.Int64 `tfsdk:"spec_revision"`
	ForceDestroy             types.Bool  `tfsdk:"force_destroy"`
	SkipDeprovisionOnDestroy types.Bool  `tfsdk:"skip_deprovision_on_destroy"`
}

type LogsModel struct {
	Storage StorageModel `tfsdk:"storage"`
}

type MetricsModel struct {
	RetentionPeriodInDays types.Int64 `tfsdk:"retention_period_in_days"`
}

type StorageModel struct {
	S3  *S3StorageModel  `tfsdk:"s3"`
	GCS *GCSStorageModel `tfsdk:"gcs"`
}

type S3StorageModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Region     types.String `tfsdk:"region"`
}

type GCSStorageModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
}

type NodeGroupsModel struct {
	Name            types.String           `tfsdk:"name"`
	NodeType        types.String           `tfsdk:"node_type"`
	CapacityPerZone types.Int64            `tfsdk:"capacity_per_zone"`
	Reservations    types.Set              `tfsdk:"reservations"`
	Zones           types.List             `tfsdk:"zones"`
	Tolerations     []TolerationModel      `tfsdk:"tolerations"`
	NodeSelector    []common.KeyValueModel `tfsdk:"selector"`
}

type NodeTypeModel struct {
	Name                  types.String  `tfsdk:"name"`
	CPUAllocatable        types.Float64 `tfsdk:"cpu_allocatable"`
	MEMAllocatableInBytes types.Float64 `tfsdk:"mem_allocatable_in_bytes"`
}

type TolerationModel struct {
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	Effect   types.String `tfsdk:"effect"`
	Operator types.String `tfsdk:"operator"`
}

type LoadBalancersModel struct {
	Public   *PublicLoadBalancerModel   `tfsdk:"public"`
	Internal *InternalLoadBalancerModel `tfsdk:"internal"`
}

type PublicLoadBalancerModel struct {
	Enabled        types.Bool             `tfsdk:"enabled"`
	SourceIPRanges []types.String         `tfsdk:"source_ip_ranges"`
	Annotations    []common.KeyValueModel `tfsdk:"annotations"`
}

type InternalLoadBalancerModel struct {
	Enabled        types.Bool             `tfsdk:"enabled"`
	SourceIPRanges []types.String         `tfsdk:"source_ip_ranges"`
	Annotations    []common.KeyValueModel `tfsdk:"annotations"`
}

func (e K8SEnvResourceModel) toSDK() (client.CreateK8SEnvInput, client.UpdateK8SEnvInput) {
	nodeGroups := nodeGroupsToSDK(e.NodeGroups)
	customNodeTypes := nodeTypesToSDK(e.CustomNodeTypes)
	loadBalancers := loadBalancersToSDK(e.LoadBalancers)
	logs := logsToSDK(e.Logs)
	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	loadBalancingStrategy := (*client.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	metrics := metricsToSDK(e.Metrics)
	distribution := client.K8SDistribution(e.Distribution.ValueString())

	create := client.CreateK8SEnvInput{
		Name: e.Name.ValueString(),
		Spec: client.CreateK8SEnvSpecInput{
			Distribution:          distribution,
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         loadBalancers,
			NodeGroups:            nodeGroups,
			CustomNodeTypes:       customNodeTypes,
			Logs:                  logs,
			Metrics:               metrics,
			MaintenanceWindows:    maintenanceWindows,
		},
	}

	strategy := client.UpdateStrategyReplace
	update := client.UpdateK8SEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: client.UpdateK8SEnvSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         loadBalancers,
			NodeGroups:            nodeGroups,
			CustomNodeTypes:       customNodeTypes,
			Logs:                  logs,
			Metrics:               metrics,
			MaintenanceWindows:    maintenanceWindows,
		},
	}

	return create, update
}

func (model *K8SEnvResourceModel) toModel(name string, spec client.K8SEnvSpecFragment) {
	model.Name = types.StringValue(name)
	model.CustomDomain = types.StringPointerValue(spec.CustomDomain)
	model.LoadBalancingStrategy = types.StringValue(string(spec.LoadBalancingStrategy))
	model.CustomNodeTypes = nodeTypesToModel(spec.CustomNodeTypes)
	model.NodeGroups = nodeGroupsToModel(spec.NodeGroups)
	model.LoadBalancers = loadBalancersToModel(spec.LoadBalancers)
	model.Logs = logsToModel(spec.Logs)
	model.MaintenanceWindows = maintenanceWindowsToModel(spec.MaintenanceWindows)
	model.Metrics = metricsToModel(spec.Metrics)
	model.Distribution = types.StringValue(string(spec.Distribution))
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *client.K8SEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *client.K8SEnvLoadBalancerPublicSpecInput
	var internal *client.K8SEnvLoadBalancerInternalSpecInput

	if loadBalancers.Public != nil {
		public = &client.K8SEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			Annotations:    common.KeyValueToSDK(loadBalancers.Public.Annotations),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
		}
	}

	if loadBalancers.Internal != nil {
		internal = &client.K8SEnvLoadBalancerInternalSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			Annotations:    common.KeyValueToSDK(loadBalancers.Internal.Annotations),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
		}
	}

	return &client.K8SEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers client.K8SEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
	model := &LoadBalancersModel{
		Public: &PublicLoadBalancerModel{
			Annotations: []common.KeyValueModel{},
			Enabled:     types.BoolValue(false),
		},
		Internal: &InternalLoadBalancerModel{
			Annotations: []common.KeyValueModel{},
			Enabled:     types.BoolValue(false),
		},
	}

	// TODO: Create helper to extract annotations and source ip ranges
	var publicAnnotations []common.KeyValueModel
	for _, a := range loadBalancers.Public.Annotations {
		publicAnnotations = append(publicAnnotations, common.KeyValueModel{
			Key:   types.StringValue(a.Key),
			Value: types.StringValue(a.Value),
		})
	}

	var publicSourceIpRanges []types.String
	for _, s := range loadBalancers.Public.SourceIPRanges {
		publicSourceIpRanges = append(publicSourceIpRanges, types.StringValue(s))
	}

	model.Public.Enabled = types.BoolValue(loadBalancers.Public.Enabled)
	model.Public.Annotations = publicAnnotations
	model.Public.SourceIPRanges = publicSourceIpRanges

	var internalAnnotations []common.KeyValueModel
	for _, element := range loadBalancers.Internal.Annotations {
		internalAnnotations = append(internalAnnotations, common.KeyValueModel{
			Key:   types.StringValue(element.Key),
			Value: types.StringValue(element.Value),
		})
	}

	var internalSourceIpRanges []types.String
	for _, s := range loadBalancers.Internal.SourceIPRanges {
		internalSourceIpRanges = append(internalSourceIpRanges, types.StringValue(s))
	}

	model.Internal.Enabled = types.BoolValue(loadBalancers.Internal.Enabled)
	model.Internal.Annotations = internalAnnotations
	model.Internal.SourceIPRanges = internalSourceIpRanges

	return model
}

func nodeGroupsToSDK(nodeGroups []NodeGroupsModel) []*client.K8SEnvNodeGroupSpecInput {
	var sdkNodeGroups []*client.K8SEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var tolerations []*client.NodeTolerationSpecInput
		for _, toleration := range np.Tolerations {
			tolerations = append(tolerations, &client.NodeTolerationSpecInput{
				Key:      toleration.Key.ValueString(),
				Value:    toleration.Value.ValueString(),
				Operator: (client.NodeTolerationOperator)(toleration.Operator.ValueString()),
				Effect:   (client.NodeTolerationEffect)(toleration.Effect.ValueString()),
			})
		}

		// TODO: Use generic helper to extract key value models
		var selector []*client.KeyValueInput
		for _, ns := range np.NodeSelector {
			selector = append(selector, &client.KeyValueInput{
				Key:   ns.Key.ValueString(),
				Value: ns.Value.ValueString(),
			})
		}

		var reservations []client.NodeReservation
		np.Reservations.ElementsAs(context.TODO(), &reservations, false)

		var zones []string
		np.Zones.ElementsAs(context.TODO(), &zones, false)

		sdkNodeGroups = append(sdkNodeGroups, &client.K8SEnvNodeGroupSpecInput{
			Name:            np.Name.ValueStringPointer(),
			NodeType:        np.NodeType.ValueString(),
			CapacityPerZone: np.CapacityPerZone.ValueInt64(),
			Zones:           zones,
			Tolerations:     tolerations,
			Selector:        selector,
			Reservations:    reservations,
		})
	}

	return sdkNodeGroups
}

func nodeGroupsToModel(nodeGroups []*client.K8SEnvSpecFragment_NodeGroups) []NodeGroupsModel {
	var modelNodeGroups []NodeGroupsModel
	for _, np := range nodeGroups {

		var tolerations []TolerationModel
		for _, toleration := range np.Tolerations {
			tolerations = append(tolerations, TolerationModel{
				Key:      types.StringValue(toleration.Key),
				Value:    types.StringValue(toleration.Value),
				Operator: types.StringValue(string(toleration.Operator)),
				Effect:   types.StringValue(string(toleration.Effect)),
			})
		}

		var nodeSelector []common.KeyValueModel
		for _, ns := range np.Selector {
			nodeSelector = append(nodeSelector, common.KeyValueModel{
				Key:   types.StringValue(ns.Key),
				Value: types.StringValue(ns.Value),
			})
		}

		modelNodeGroups = append(modelNodeGroups, NodeGroupsModel{
			Name:            types.StringValue(np.Name),
			NodeType:        types.StringValue(np.NodeType),
			CapacityPerZone: types.Int64Value(np.CapacityPerZone),
			Zones:           common.ListToModel(np.Zones),
			Reservations:    common.ReservationsToModel(np.Reservations),
			Tolerations:     tolerations,
			NodeSelector:    nodeSelector,
		})
	}

	return modelNodeGroups
}

func nodeTypesToSDK(nodeTypes []NodeTypeModel) []*client.K8SEnvCustomNodeTypeSpecInput {
	var sdkNodeTypes []*client.K8SEnvCustomNodeTypeSpecInput
	for _, nt := range nodeTypes {
		sdkNodeTypes = append(sdkNodeTypes, &client.K8SEnvCustomNodeTypeSpecInput{
			Name:                  nt.Name.ValueString(),
			CPUAllocatable:        nt.CPUAllocatable.ValueFloat64(),
			MemAllocatableInBytes: nt.MEMAllocatableInBytes.ValueFloat64(),
		})
	}

	return sdkNodeTypes
}

func nodeTypesToModel(nodeTypes []*client.K8SEnvSpecFragment_CustomNodeTypes) []NodeTypeModel {
	var modelNodeTypes []NodeTypeModel
	for _, nt := range nodeTypes {
		modelNodeTypes = append(modelNodeTypes, NodeTypeModel{
			Name:                  types.StringValue(nt.Name),
			CPUAllocatable:        types.Float64Value(nt.CPUAllocatable),
			MEMAllocatableInBytes: types.Float64Value(float64(nt.MemAllocatableInBytes)),
		})
	}

	return modelNodeTypes
}

func logsToSDK(logs *LogsModel) *client.K8SEnvLogsSpecInput {
	if logs == nil {
		return nil
	}

	var model = &client.K8SEnvLogsSpecInput{
		Storage: &client.K8SEnvSpecLogsStorageSpecInput{},
	}
	if logs.Storage.S3 != nil {
		model.Storage.S3 = &client.K8SEnvSpecLogsStorageS3SpecInput{
			BucketName: logs.Storage.S3.BucketName.ValueStringPointer(),
			Region:     logs.Storage.S3.Region.ValueStringPointer(),
		}
	}

	if logs.Storage.GCS != nil {
		model.Storage.Gcs = &client.K8SEnvSpecLogsStorageGCSSpecInput{
			BucketName: logs.Storage.GCS.BucketName.ValueStringPointer(),
		}
	}

	return model
}

func logsToModel(logs client.K8SEnvSpecFragment_Logs) *LogsModel {
	var model = &LogsModel{Storage: StorageModel{}}
	if logs.Storage.S3 != nil {
		model.Storage.S3 = &S3StorageModel{
			BucketName: types.StringValue(*logs.Storage.S3.BucketName),
			Region:     types.StringValue(*logs.Storage.S3.Region),
		}
	}

	if logs.Storage.Gcs != nil {
		model.Storage.GCS = &GCSStorageModel{
			BucketName: types.StringValue(*logs.Storage.S3.BucketName),
		}
	}

	return model
}

func metricsToSDK(metrics *MetricsModel) *client.K8SEnvMetricsSpecInput {
	if metrics == nil {
		return nil
	}

	return &client.K8SEnvMetricsSpecInput{
		RetentionPeriodInDays: metrics.RetentionPeriodInDays.ValueInt64Pointer(),
	}
}

func metricsToModel(metrics client.K8SEnvSpecFragment_Metrics) *MetricsModel {
	return &MetricsModel{
		RetentionPeriodInDays: types.Int64PointerValue(metrics.RetentionPeriodInDays),
	}
}

func maintenanceWindowsToModel(input []*client.K8SEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
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
