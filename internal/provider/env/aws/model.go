package env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AWSEnvResourceModel struct {
	Id                           types.String                    `tfsdk:"id"`
	Name                         types.String                    `tfsdk:"name"`
	CustomDomain                 types.String                    `tfsdk:"custom_domain"`
	LoadBalancingStrategy        types.String                    `tfsdk:"load_balancing_strategy"`
	Region                       types.String                    `tfsdk:"region"`
	PermissionsBoundaryPolicyArn types.String                    `tfsdk:"permissions_boundary_policy_arn"`
	ResourcePrefix               types.String                    `tfsdk:"resource_prefix"`
	NAT                          types.Bool                      `tfsdk:"nat"`
	CIDR                         types.String                    `tfsdk:"cidr"`
	AWSAccountID                 types.String                    `tfsdk:"aws_account_id"`
	Zones                        types.List                      `tfsdk:"zones"`
	LoadBalancers                *LoadBalancersModel             `tfsdk:"load_balancers"`
	NodeGroups                   []common.NodeGroupsModel        `tfsdk:"node_groups"`
	PeeringConnections           []AWSEnvPeeringConnectionModel  `tfsdk:"peering_connections"`
	Endpoints                    []AWSEnvEndpointModel           `tfsdk:"endpoints"`
	Tags                         []common.KeyValueModel          `tfsdk:"tags"`
	CloudConnect                 types.Bool                      `tfsdk:"cloud_connect"`
	MaintenanceWindows           []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	ExternalBuckets              []AWSEnvExternalBucketModel     `tfsdk:"external_buckets"`
	Backups                      *AWSEnvBackupsModel             `tfsdk:"backups"`
	Iceberg                      *AWSEnvIcebergModel             `tfsdk:"iceberg"`
	// MetricsEndpoint              *AWSEnvMetricsEndpointModel     `tfsdk:"metrics_endpoint"`
	EksLogging                   types.Bool                      `tfsdk:"eks_logging"`

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
	CrossZone      types.Bool     `tfsdk:"cross_zone"`
}

type InternalLoadBalancerModel struct {
	Enabled                          types.Bool     `tfsdk:"enabled"`
	SourceIPRanges                   []types.String `tfsdk:"source_ip_ranges"`
	CrossZone                        types.Bool     `tfsdk:"cross_zone"`
	EndpointServiceAllowedPrincipals []types.String `tfsdk:"endpoint_service_allowed_principals"`
	EndpointServiceSupportedRegions  []types.String `tfsdk:"endpoint_service_supported_regions"`
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

type AWSEnvExternalBucketModel struct {
	Name types.String `tfsdk:"name"`
}

type AWSEnvBackupsModel struct {
	CustomBucket *AWSEnvCustomBucketModel `tfsdk:"custom_bucket"`
}

type AWSEnvCustomBucketModel struct {
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	RoleArn types.String `tfsdk:"role_arn"`
}

type AWSEnvIcebergModel struct {
	Catalogs []AWSEnvIcebergCatalogModel `tfsdk:"catalogs"`
}

type AWSEnvIcebergCatalogModel struct {
	Name                   types.String                          `tfsdk:"name"`
	Type                   types.String                          `tfsdk:"type"`
	CustomS3Bucket         types.String                          `tfsdk:"custom_s3_bucket"`
	CustomS3BucketPath     types.String                          `tfsdk:"custom_s3_bucket_path"`
	CustomS3TableBucketARN types.String                          `tfsdk:"custom_s3_table_bucket_arn"`
	AWSRegion              types.String                          `tfsdk:"aws_region"`
	AnonymousAccessEnabled types.Bool                            `tfsdk:"anonymous_access_enabled"`
	Maintenance            *AWSEnvIcebergCatalogMaintenanceModel `tfsdk:"maintenance"`
	Watches                []AWSEnvIcebergCatalogWatchModel      `tfsdk:"watches"`
	RoleARN                types.String                          `tfsdk:"role_arn"`
	AssumeRoleARNRW        types.String                          `tfsdk:"assume_role_arn_rw"`
	AssumeRoleARNRO        types.String                          `tfsdk:"assume_role_arn_ro"`
}

type AWSEnvIcebergCatalogMaintenanceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type AWSEnvIcebergCatalogWatchModel struct {
	Table                        types.String   `tfsdk:"table"`
	PathsRelativeToTableLocation []types.String `tfsdk:"paths_relative_to_table_location"`
}

type AWSEnvMetricsEndpointModel struct {
	Enabled        types.Bool     `tfsdk:"enabled"`
	SourceIPRanges []types.String `tfsdk:"source_ip_ranges"`
}

func (e AWSEnvResourceModel) toSDK() (sdk.CreateAWSEnvInput, sdk.UpdateAWSEnvInput) {
	var zones []string
	e.Zones.ElementsAs(context.TODO(), &zones, false)

	var peeringConnections []*sdk.AWSEnvPeeringConnectionSpecInput
	for _, p := range e.PeeringConnections {
		peeringConnections = append(peeringConnections, &sdk.AWSEnvPeeringConnectionSpecInput{
			AWSAccountID: p.AWSAccountID.ValueStringPointer(),
			VpcID:        p.VpcID.ValueString(),
			VpcRegion:    p.VpcRegion.ValueStringPointer(),
		})
	}

	var endpoints []*sdk.AWSEnvEndpointSpecInput
	for _, e := range e.Endpoints {
		endpoints = append(endpoints, &sdk.AWSEnvEndpointSpecInput{
			ServiceName: e.ServiceName.ValueString(),
			Alias:       e.Alias.ValueStringPointer(),
			PrivateDNS:  e.PrivateDNS.ValueBoolPointer(),
		})
	}

	var tags []*sdk.KeyValueInput
	for _, t := range e.Tags {
		tags = append(tags, &sdk.KeyValueInput{
			Key:   t.Key.ValueString(),
			Value: t.Value.ValueString(),
		})
	}

	var externalBuckets []*sdk.AWSEnvExternalBucketSpecInput
	for _, b := range e.ExternalBuckets {
		externalBuckets = append(externalBuckets, &sdk.AWSEnvExternalBucketSpecInput{
			Name: b.Name.ValueString(),
		})
	}

	backups := backupsToSDK(e.Backups)
	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	LoadBalancers := loadBalancersToSDK(e.LoadBalancers)
	nodeGroups := nodeGroupsToSDK(e.NodeGroups)
	loadBalancingStrategy := (*sdk.LoadBalancingStrategy)(e.LoadBalancingStrategy.ValueStringPointer())
	cloudConnect := e.CloudConnect.ValueBool()

	iceberg := icebergToSDK(e.Iceberg)
	// metricsEndpoint := metricsEndpointToSDK(e.MetricsEndpoint)

	create := sdk.CreateAWSEnvInput{
		Name: e.Name.ValueString(),
		Spec: &sdk.CreateAWSEnvSpecInput{
			CustomDomain:                 e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy:        loadBalancingStrategy,
			LoadBalancers:                LoadBalancers,
			NodeGroups:                   nodeGroups,
			Region:                       e.Region.ValueString(),
			Nat:                          e.NAT.ValueBoolPointer(),
			AWSAccountID:                 e.AWSAccountID.ValueString(),
			Cidr:                         e.CIDR.ValueString(),
			Zones:                        zones,
			PeeringConnections:           peeringConnections,
			Endpoints:                    endpoints,
			Tags:                         tags,
			CloudConnect:                 &cloudConnect,
			MaintenanceWindows:           maintenanceWindows,
			PermissionsBoundaryPolicyArn: e.PermissionsBoundaryPolicyArn.ValueStringPointer(),
			ResourcePrefix:               e.ResourcePrefix.ValueStringPointer(),
			ExternalBuckets:              externalBuckets,
			Backups:                      backups,
			Iceberg:                      iceberg,
			MetricsEndpoint:              nil, // metricsEndpoint
			EksLogging:                   e.EksLogging.ValueBoolPointer(),
		},
	}

	icebergUpdate := icebergToUpdateSDK(e.Iceberg)

	strategy := sdk.UpdateStrategyReplace
	update := sdk.UpdateAWSEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &sdk.AWSEnvUpdateSpecInput{
			CustomDomain:          e.CustomDomain.ValueStringPointer(),
			LoadBalancingStrategy: loadBalancingStrategy,
			LoadBalancers:         LoadBalancers,
			NodeGroups:            nodeGroups,
			Zones:                 zones,
			PeeringConnections:    peeringConnections,
			Endpoints:             endpoints,
			Tags:                  tags,
			MaintenanceWindows:    maintenanceWindows,
			ExternalBuckets:       externalBuckets,
			Backups:               backups,
			Iceberg:               icebergUpdate,
			MetricsEndpoint:       nil, // metricsEndpoint
			EksLogging:            e.EksLogging.ValueBoolPointer(),
		},
	}

	return create, update
}

func (model *AWSEnvResourceModel) toModel(env sdk.GetAWSEnv_AWSEnv) {
	model.Name = types.StringValue(env.Name)
	model.CIDR = types.StringValue(env.Spec.Cidr)
	model.Region = types.StringValue(env.Spec.Region)
	model.NAT = types.BoolValue(env.Spec.Nat)
	model.AWSAccountID = types.StringValue(env.Spec.AWSAccountID)
	model.CustomDomain = types.StringPointerValue(env.Spec.CustomDomain)
	model.LoadBalancingStrategy = types.StringValue(string(env.Spec.LoadBalancingStrategy))
	model.LoadBalancers = loadBalancersToModel(env.Spec.LoadBalancers)
	model.NodeGroups = nodeGroupsToModel(env.Spec.NodeGroups)
	model.MaintenanceWindows = maintenanceWindowsToModel(env.Spec.MaintenanceWindows)
	model.Zones = common.ListToModel(env.Spec.Zones)
	model.PermissionsBoundaryPolicyArn = types.StringPointerValue(env.Spec.PermissionsBoundaryPolicyArn)
	model.ResourcePrefix = types.StringValue(env.Spec.ResourcePrefix)

	var peeringConnections []AWSEnvPeeringConnectionModel
	for _, p := range env.Spec.PeeringConnections {
		peeringConnections = append(peeringConnections, AWSEnvPeeringConnectionModel{
			AWSAccountID: types.StringPointerValue(p.AWSAccountID),
			VpcID:        types.StringValue(p.VpcID),
			VpcRegion:    types.StringPointerValue(p.VpcRegion),
		})
	}

	var endpoints []AWSEnvEndpointModel
	for _, e := range env.Spec.Endpoints {
		endpoints = append(endpoints, AWSEnvEndpointModel{
			ServiceName: types.StringValue(e.ServiceName),
			Alias:       types.StringPointerValue(e.Alias),
			PrivateDNS:  types.BoolValue(e.PrivateDNS),
		})
	}

	var tags []common.KeyValueModel
	for _, t := range env.Spec.Tags {
		tags = append(tags, common.KeyValueModel{
			Key:   types.StringValue(t.Key),
			Value: types.StringValue(t.Value),
		})
	}

	var externalBuckets []AWSEnvExternalBucketModel
	for _, b := range env.Spec.ExternalBuckets {
		externalBuckets = append(externalBuckets, AWSEnvExternalBucketModel{
			Name: types.StringValue(b.Name),
		})
	}

	backups := backupsToModel(env.Spec.Backups)
	iceberg := icebergToModel(env.Spec.Iceberg)

	model.Tags = tags
	model.Endpoints = endpoints
	model.ExternalBuckets = externalBuckets
	model.Backups = backups
	model.Iceberg = iceberg
	model.PeeringConnections = peeringConnections
	model.SpecRevision = types.Int64Value(env.SpecRevision)
	model.CloudConnect = types.BoolValue(env.Spec.CloudConnect)
	model.EksLogging = types.BoolValue(env.Spec.EksLogging)
	// model.MetricsEndpoint = metricsEndpointToModel(&env.Spec.MetricsEndpoint)
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

		var endpointServiceSupportedRegions []string
		for _, sr := range loadBalancers.Internal.EndpointServiceSupportedRegions {
			endpointServiceSupportedRegions = append(endpointServiceSupportedRegions, sr.ValueString())
		}

		internal = &sdk.AWSEnvLoadBalancerInternalSpecInput{
			Enabled:                          loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges:                   common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
			CrossZone:                        loadBalancers.Internal.CrossZone.ValueBoolPointer(),
			EndpointServiceAllowedPrincipals: endpointServiceAllowedPrincipals,
			EndpointServiceSupportedRegions:  endpointServiceSupportedRegions,
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

	var endpointServiceSupportedRegions []types.String
	for _, e := range loadBalancers.Internal.EndpointServiceSupportedRegions {
		endpointServiceSupportedRegions = append(endpointServiceSupportedRegions, types.StringValue(e))
	}

	model.Internal = &InternalLoadBalancerModel{
		Enabled:                          types.BoolValue(loadBalancers.Internal.Enabled),
		SourceIPRanges:                   internalSourceIpRanges,
		CrossZone:                        types.BoolValue(loadBalancers.Internal.CrossZone),
		EndpointServiceAllowedPrincipals: endpointServiceAllowedPrincipals,
		EndpointServiceSupportedRegions:  endpointServiceSupportedRegions,
	}

	return model
}

func nodeGroupsToSDK(nodeGroups []common.NodeGroupsModel) []*sdk.AWSEnvNodeGroupSpecInput {
	var sdkNodeGroups []*sdk.AWSEnvNodeGroupSpecInput
	for _, np := range nodeGroups {
		var reservations []sdk.NodeReservation
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

func reorderNodeGroups(model []common.NodeGroupsModel, nodeGroups []*sdk.AWSEnvSpecFragment_NodeGroups) []*sdk.AWSEnvSpecFragment_NodeGroups {
	orderedNodeGroups := make([]*sdk.AWSEnvSpecFragment_NodeGroups, 0, len(nodeGroups))
	usedNodeGroups := make(map[string]bool)

	// First, add node groups that exist in the model in the correct order
	for _, ng := range model {
		for _, apiGroup := range nodeGroups {
			if ng.NodeType.ValueString() == apiGroup.NodeType {
				orderedNodeGroups = append(orderedNodeGroups, apiGroup)
				usedNodeGroups[apiGroup.NodeType] = true
				break
			}
		}
	}

	// Then, add any remaining node groups from the API that weren't in the model
	for _, apiGroup := range nodeGroups {
		if !usedNodeGroups[apiGroup.NodeType] {
			orderedNodeGroups = append(orderedNodeGroups, apiGroup)
		}
	}

	return orderedNodeGroups
}

func backupsToSDK(backups *AWSEnvBackupsModel) *sdk.AWSEnvBackupsSpecInput {
	if backups == nil || backups.CustomBucket == nil {
		return nil
	}

	return &sdk.AWSEnvBackupsSpecInput{
		CustomBucket: &sdk.AWSEnvBackupsCustomBucketSpecInput{
			Name:    backups.CustomBucket.Name.ValueString(),
			Region:  backups.CustomBucket.Region.ValueString(),
			RoleArn: backups.CustomBucket.RoleArn.ValueString(),
		},
	}
}

func backupsToModel(backups *sdk.AWSEnvSpecFragment_Backups) *AWSEnvBackupsModel {
	if backups == nil || backups.CustomBucket == nil {
		return nil
	}

	return &AWSEnvBackupsModel{
		CustomBucket: &AWSEnvCustomBucketModel{
			Name:    types.StringValue(backups.CustomBucket.Name),
			Region:  types.StringValue(backups.CustomBucket.Region),
			RoleArn: types.StringValue(backups.CustomBucket.RoleArn),
		},
	}
}

func icebergToSDK(iceberg *AWSEnvIcebergModel) *sdk.IcebergInputSpec {
	if iceberg == nil {
		return nil
	}

	var catalogs []*sdk.IcebergCatalogInputSpec
	for _, c := range iceberg.Catalogs {
		catalog := &sdk.IcebergCatalogInputSpec{
			Name:                   c.Name.ValueStringPointer(),
			Type:                   sdk.IcebergCatalogTypeSpec(c.Type.ValueString()),
			CustomS3Bucket:         c.CustomS3Bucket.ValueStringPointer(),
			CustomS3BucketPath:     c.CustomS3BucketPath.ValueStringPointer(),
			CustomS3TableBucketArn: c.CustomS3TableBucketARN.ValueStringPointer(),
			AWSRegion:              c.AWSRegion.ValueStringPointer(),
			AnonymousAccessEnabled: c.AnonymousAccessEnabled.ValueBoolPointer(),
			RoleArn:                c.RoleARN.ValueStringPointer(),
			AssumeRoleArnrw:        c.AssumeRoleARNRW.ValueStringPointer(),
			AssumeRoleArnro:        c.AssumeRoleARNRO.ValueStringPointer(),
		}

		if c.Maintenance != nil {
			catalog.Maintenance = &sdk.IcebergCatalogMaintenanceInputSpec{
				Enabled: c.Maintenance.Enabled.ValueBool(),
			}
		}

		var watches []*sdk.IcebergCatalogWatchInputSpec
		for _, w := range c.Watches {
			var paths []string
			for _, p := range w.PathsRelativeToTableLocation {
				paths = append(paths, p.ValueString())
			}
			watches = append(watches, &sdk.IcebergCatalogWatchInputSpec{
				Table:                        w.Table.ValueString(),
				PathsRelativeToTableLocation: paths,
			})
		}
		catalog.Watches = watches

		catalogs = append(catalogs, catalog)
	}

	return &sdk.IcebergInputSpec{
		Catalogs: catalogs,
	}
}

func icebergToUpdateSDK(iceberg *AWSEnvIcebergModel) *sdk.IcebergUpdateInputSpec {
	if iceberg == nil {
		return nil
	}

	var catalogs []*sdk.IcebergCatalogInputSpec
	for _, c := range iceberg.Catalogs {
		catalog := &sdk.IcebergCatalogInputSpec{
			Name:                   c.Name.ValueStringPointer(),
			Type:                   sdk.IcebergCatalogTypeSpec(c.Type.ValueString()),
			CustomS3Bucket:         c.CustomS3Bucket.ValueStringPointer(),
			CustomS3BucketPath:     c.CustomS3BucketPath.ValueStringPointer(),
			CustomS3TableBucketArn: c.CustomS3TableBucketARN.ValueStringPointer(),
			AWSRegion:              c.AWSRegion.ValueStringPointer(),
			AnonymousAccessEnabled: c.AnonymousAccessEnabled.ValueBoolPointer(),
			RoleArn:                c.RoleARN.ValueStringPointer(),
			AssumeRoleArnrw:        c.AssumeRoleARNRW.ValueStringPointer(),
			AssumeRoleArnro:        c.AssumeRoleARNRO.ValueStringPointer(),
		}

		if c.Maintenance != nil {
			catalog.Maintenance = &sdk.IcebergCatalogMaintenanceInputSpec{
				Enabled: c.Maintenance.Enabled.ValueBool(),
			}
		}

		var watches []*sdk.IcebergCatalogWatchInputSpec
		for _, w := range c.Watches {
			var paths []string
			for _, p := range w.PathsRelativeToTableLocation {
				paths = append(paths, p.ValueString())
			}
			watches = append(watches, &sdk.IcebergCatalogWatchInputSpec{
				Table:                        w.Table.ValueString(),
				PathsRelativeToTableLocation: paths,
			})
		}
		catalog.Watches = watches

		catalogs = append(catalogs, catalog)
	}

	return &sdk.IcebergUpdateInputSpec{
		Catalogs: catalogs,
	}
}

func icebergToModel(iceberg *sdk.AWSEnvSpecFragment_Iceberg) *AWSEnvIcebergModel {
	if iceberg == nil || len(iceberg.Catalogs) == 0 {
		return nil
	}

	var catalogs []AWSEnvIcebergCatalogModel
	for _, c := range iceberg.Catalogs {
		catalog := AWSEnvIcebergCatalogModel{
			Name:                   types.StringPointerValue(c.Name),
			Type:                   types.StringValue(string(c.Type)),
			CustomS3Bucket:         types.StringPointerValue(c.CustomS3Bucket),
			CustomS3BucketPath:     types.StringPointerValue(c.CustomS3BucketPath),
			CustomS3TableBucketARN: types.StringPointerValue(c.CustomS3TableBucketArn),
			AWSRegion:              types.StringPointerValue(c.AWSRegion),
			AnonymousAccessEnabled: types.BoolPointerValue(c.AnonymousAccessEnabled),
			RoleARN:                types.StringPointerValue(c.RoleArn),
			AssumeRoleARNRW:        types.StringPointerValue(c.AssumeRoleArnrw),
			AssumeRoleARNRO:        types.StringPointerValue(c.AssumeRoleArnro),
		}

		catalog.Maintenance = &AWSEnvIcebergCatalogMaintenanceModel{
			Enabled: types.BoolValue(c.Maintenance.Enabled),
		}

		var watches []AWSEnvIcebergCatalogWatchModel
		for _, w := range c.Watches {
			var paths []types.String
			for _, p := range w.PathsRelativeToTableLocation {
				paths = append(paths, types.StringValue(p))
			}
			watches = append(watches, AWSEnvIcebergCatalogWatchModel{
				Table:                        types.StringValue(w.Table),
				PathsRelativeToTableLocation: paths,
			})
		}
		catalog.Watches = watches

		catalogs = append(catalogs, catalog)
	}

	return &AWSEnvIcebergModel{
		Catalogs: catalogs,
	}
}

func metricsEndpointToSDK(endpoint *AWSEnvMetricsEndpointModel) *sdk.MetricsEndpointSpecInput {
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

func metricsEndpointToModel(endpoint *sdk.AWSEnvSpecFragment_MetricsEndpoint) *AWSEnvMetricsEndpointModel {
	if endpoint == nil {
		return nil
	}

	var sourceIPRanges []types.String
	for _, ip := range endpoint.SourceIPRanges {
		sourceIPRanges = append(sourceIPRanges, types.StringValue(ip))
	}

	return &AWSEnvMetricsEndpointModel{
		Enabled:        types.BoolValue(endpoint.Enabled),
		SourceIPRanges: sourceIPRanges,
	}
}

func reorderTags(model []common.KeyValueModel, tags []*sdk.AWSEnvSpecFragment_Tags) []*sdk.AWSEnvSpecFragment_Tags {
	orderedTags := make([]*sdk.AWSEnvSpecFragment_Tags, 0, len(tags))
	usedTags := make(map[string]bool)

	// First, add tags that exist in the model in the correct order
	for _, tag := range model {
		for _, apiTag := range tags {
			if tag.Key.ValueString() == apiTag.Key {
				orderedTags = append(orderedTags, apiTag)
				usedTags[apiTag.Key] = true
				break
			}
		}
	}

	// Then, add any remaining tags from the API that weren't in the model
	for _, apiTag := range tags {
		if !usedTags[apiTag.Key] {
			orderedTags = append(orderedTags, apiTag)
		}
	}

	return orderedTags
}
