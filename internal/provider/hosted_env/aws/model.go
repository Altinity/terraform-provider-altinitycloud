package hosted_env

import (
	"context"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	hosted "github.com/altinity/terraform-provider-altinitycloud/internal/provider/hosted_env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostedAWSEnvResourceModel struct {
	Id                 types.String                    `tfsdk:"id"`
	Name               types.String                    `tfsdk:"name"`
	Region             types.String                    `tfsdk:"region"`
	CIDR               types.String                    `tfsdk:"cidr"`
	ZoneIDs            types.List                      `tfsdk:"zone_ids"`
	ResourcePrefix     types.String                    `tfsdk:"resource_prefix"`
	KmsKeyArn          types.String                    `tfsdk:"kms_key_arn"`
	CustomDomains      types.List                      `tfsdk:"custom_domains"`
	LoadBalancers      *LoadBalancersModel             `tfsdk:"load_balancers"`
	NodeGroups         []hosted.NodeGroupsModel        `tfsdk:"node_groups"`
	MaintenanceWindows []common.MaintenanceWindowModel `tfsdk:"maintenance_windows"`
	Endpoints          []EndpointModel                 `tfsdk:"endpoints"`
	ExternalBuckets    []ExternalBucketModel           `tfsdk:"external_buckets"`
	Backups            *BackupsModel                   `tfsdk:"backups"`
	Iceberg            *IcebergModel                   `tfsdk:"iceberg"`
	MetricsEndpoint    *hosted.MetricsEndpointModel    `tfsdk:"metrics_endpoint"`
	Datadog            *common.DatadogModel            `tfsdk:"datadog"`

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
	Enabled                          types.Bool     `tfsdk:"enabled"`
	SourceIPRanges                   []types.String `tfsdk:"source_ip_ranges"`
	EndpointServiceAllowedPrincipals []types.String `tfsdk:"endpoint_service_allowed_principals"`
	EndpointServiceSupportedRegions  []types.String `tfsdk:"endpoint_service_supported_regions"`
}

type EndpointModel struct {
	ServiceName types.String `tfsdk:"service_name"`
	Alias       types.String `tfsdk:"alias"`
}

type ExternalBucketModel struct {
	Name      types.String `tfsdk:"name"`
	KmsKeyArn types.String `tfsdk:"kms_key_arn"`
}

type BackupsModel struct {
	CustomBucket *CustomBucketModel `tfsdk:"custom_bucket"`
}

type CustomBucketModel struct {
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	RoleArn types.String `tfsdk:"role_arn"`
}

type IcebergModel struct {
	Catalogs []IcebergCatalogModel `tfsdk:"catalogs"`
}

type IcebergCatalogModel struct {
	Name                   types.String                    `tfsdk:"name"`
	Type                   types.String                    `tfsdk:"type"`
	CustomS3Bucket         types.String                    `tfsdk:"custom_s3_bucket"`
	CustomS3BucketPath     types.String                    `tfsdk:"custom_s3_bucket_path"`
	CustomS3TableBucketArn types.String                    `tfsdk:"custom_s3_table_bucket_arn"`
	Region                 types.String                    `tfsdk:"region"`
	AnonymousAccessEnabled types.Bool                      `tfsdk:"anonymous_access_enabled"`
	Maintenance            *IcebergCatalogMaintenanceModel `tfsdk:"maintenance"`
	Watches                []IcebergCatalogWatchModel      `tfsdk:"watches"`
}

type IcebergCatalogMaintenanceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type IcebergCatalogWatchModel struct {
	Table                        types.String   `tfsdk:"table"`
	PathsRelativeToTableLocation []types.String `tfsdk:"paths_relative_to_table_location"`
}

func (e HostedAWSEnvResourceModel) toSDK(ctx context.Context) (sdk.CreateHostedAWSEnvInput, sdk.UpdateHostedAWSEnvInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	var zoneIDs []string
	if !e.ZoneIDs.IsUnknown() && !e.ZoneIDs.IsNull() {
		allDiags.Append(e.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
	}

	var customDomains []string
	if !e.CustomDomains.IsUnknown() && !e.CustomDomains.IsNull() {
		allDiags.Append(e.CustomDomains.ElementsAs(ctx, &customDomains, false)...)
	}

	var endpoints []*sdk.HostedAWSEnvEndpointSpecInput
	for _, endpoint := range e.Endpoints {
		endpoints = append(endpoints, &sdk.HostedAWSEnvEndpointSpecInput{
			ServiceName: endpoint.ServiceName.ValueString(),
			Alias:       endpoint.Alias.ValueStringPointer(),
		})
	}

	var externalBuckets []*sdk.HostedAWSEnvExternalBucketSpecInput
	for _, bucket := range e.ExternalBuckets {
		externalBuckets = append(externalBuckets, &sdk.HostedAWSEnvExternalBucketSpecInput{
			Name:      bucket.Name.ValueString(),
			KmsKeyArn: bucket.KmsKeyArn.ValueStringPointer(),
		})
	}

	nodeGroups, diags := nodeGroupsToSDK(ctx, e.NodeGroups)
	allDiags.Append(diags...)

	loadBalancers := loadBalancersToSDK(e.LoadBalancers)
	maintenanceWindows := common.MaintenanceWindowsToSDK(e.MaintenanceWindows)
	backups := backupsToSDK(e.Backups)
	metricsEndpoint := hosted.MetricsEndpointToSDK(e.MetricsEndpoint)
	datadog := common.DatadogToSDK(e.Datadog)
	catalogs := icebergCatalogsToSDK(e.Iceberg)

	var iceberg *sdk.HostedAWSEnvIcebergInputSpec
	var icebergUpdate *sdk.HostedAWSEnvIcebergUpdateInputSpec
	if e.Iceberg != nil {
		iceberg = &sdk.HostedAWSEnvIcebergInputSpec{Catalogs: catalogs}
		icebergUpdate = &sdk.HostedAWSEnvIcebergUpdateInputSpec{Catalogs: catalogs}
	}

	create := sdk.CreateHostedAWSEnvInput{
		Name: e.Name.ValueString(),
		Spec: &sdk.CreateHostedAWSEnvSpecInput{
			Region:             e.Region.ValueString(),
			ZoneIDs:            zoneIDs,
			ResourcePrefix:     e.ResourcePrefix.ValueStringPointer(),
			KmsKeyArn:          e.KmsKeyArn.ValueStringPointer(),
			CustomDomains:      customDomains,
			LoadBalancers:      loadBalancers,
			NodeGroups:         nodeGroups,
			MaintenanceWindows: maintenanceWindows,
			Endpoints:          endpoints,
			ExternalBuckets:    externalBuckets,
			Backups:            backups,
			Iceberg:            iceberg,
			MetricsEndpoint:    metricsEndpoint,
			Datadog:            datadog,
		},
	}

	strategy := sdk.UpdateStrategyReplace
	update := sdk.UpdateHostedAWSEnvInput{
		Name:           e.Name.ValueString(),
		UpdateStrategy: &strategy,
		Spec: &sdk.HostedAWSEnvUpdateSpecInput{
			ZoneIDs:            zoneIDs,
			CustomDomains:      customDomains,
			LoadBalancers:      loadBalancers,
			NodeGroups:         nodeGroups,
			MaintenanceWindows: maintenanceWindows,
			Endpoints:          endpoints,
			ExternalBuckets:    externalBuckets,
			Backups:            backups,
			Iceberg:            icebergUpdate,
			MetricsEndpoint:    metricsEndpoint,
			Datadog:            datadog,
		},
	}

	return create, update, allDiags
}

// applySpec maps the spec every create/read/update returns, after reordering the
// lists the API may hand back in a different order than the user configured them.
func (model *HostedAWSEnvResourceModel) applySpec(ctx context.Context, name string, spec *sdk.HostedAWSEnvSpecFragment, specRevision int64) diag.Diagnostics {
	var allDiags diag.Diagnostics
	if spec == nil {
		allDiags.AddError("Empty environment spec", "The API returned an environment without a spec.")
		return allDiags
	}

	nodeGroupSpecKey := func(s *sdk.HostedAWSEnvSpecFragment_NodeGroups) string { return s.Name }
	spec.NodeGroups = common.ReorderByKey(model.NodeGroups, spec.NodeGroups, hosted.NodeGroupKey, nodeGroupSpecKey)
	allDiags.Append(common.ReorderNodeGroupZones(ctx, model.NodeGroups, spec.NodeGroups,
		hosted.NodeGroupKey,
		nodeGroupSpecKey,
		func(m hosted.NodeGroupsModel) types.List { return m.ZoneIDs },
		func(s *sdk.HostedAWSEnvSpecFragment_NodeGroups) *[]string { return &s.ZoneIDs },
	)...)
	spec.MaintenanceWindows = common.ReorderByKey(model.MaintenanceWindows, spec.MaintenanceWindows,
		func(m common.MaintenanceWindowModel) string { return m.Name.ValueString() },
		func(s *sdk.HostedAWSEnvSpecFragment_MaintenanceWindows) string { return s.Name },
	)
	spec.Endpoints = common.ReorderByKey(model.Endpoints, spec.Endpoints,
		func(m EndpointModel) string { return m.ServiceName.ValueString() },
		func(s *sdk.HostedAWSEnvSpecFragment_Endpoints) string { return s.ServiceName },
	)
	spec.ExternalBuckets = common.ReorderByKey(model.ExternalBuckets, spec.ExternalBuckets,
		func(m ExternalBucketModel) string { return m.Name.ValueString() },
		func(s *sdk.HostedAWSEnvSpecFragment_ExternalBuckets) string { return s.Name },
	)
	reorderIceberg(model.Iceberg, spec.Iceberg)

	zoneIDs, diags := common.ReorderList(ctx, model.ZoneIDs, spec.ZoneIDs)
	allDiags.Append(diags...)

	model.Name = types.StringValue(name)
	model.SpecRevision = types.Int64Value(specRevision)
	model.Region = types.StringValue(spec.Region)
	model.CIDR = types.StringValue(spec.Cidr)
	model.ResourcePrefix = types.StringValue(spec.ResourcePrefix)
	model.KmsKeyArn = types.StringPointerValue(spec.KmsKeyArn)

	model.ZoneIDs, diags = common.ListToModel(zoneIDs)
	allDiags.Append(diags...)
	model.CustomDomains, diags = hosted.CustomDomainsToModel(model.CustomDomains, spec.CustomDomains)
	allDiags.Append(diags...)
	model.NodeGroups, diags = nodeGroupsToModel(spec.NodeGroups)
	allDiags.Append(diags...)

	model.LoadBalancers = loadBalancersToModel(spec.LoadBalancers)
	model.MaintenanceWindows = maintenanceWindowsToModel(spec.MaintenanceWindows)
	model.Endpoints = endpointsToModel(spec.Endpoints)
	model.ExternalBuckets = externalBucketsToModel(spec.ExternalBuckets)
	model.Backups = backupsToModel(spec.Backups)
	model.Iceberg = icebergToModel(spec.Iceberg)
	model.MetricsEndpoint = hosted.MetricsEndpointToModel(model.MetricsEndpoint, spec.MetricsEndpoint.Enabled, spec.MetricsEndpoint.SourceIPRanges)
	model.Datadog = hosted.DatadogToModel(model.Datadog, spec.Datadog.Enabled, spec.Datadog.Domain, spec.Datadog.LogsEnabled, spec.Datadog.MetricsEnabled)

	return allDiags
}

func loadBalancersToSDK(loadBalancers *LoadBalancersModel) *sdk.HostedAWSEnvLoadBalancersSpecInput {
	if loadBalancers == nil {
		return nil
	}

	var public *sdk.HostedAWSEnvLoadBalancerPublicSpecInput
	if loadBalancers.Public != nil {
		public = &sdk.HostedAWSEnvLoadBalancerPublicSpecInput{
			Enabled:        loadBalancers.Public.Enabled.ValueBoolPointer(),
			SourceIPRanges: common.ListStringToSDK(loadBalancers.Public.SourceIPRanges),
		}
	}

	var internal *sdk.HostedAWSEnvLoadBalancerInternalSpecInput
	if loadBalancers.Internal != nil {
		internal = &sdk.HostedAWSEnvLoadBalancerInternalSpecInput{
			Enabled:                          loadBalancers.Internal.Enabled.ValueBoolPointer(),
			SourceIPRanges:                   common.ListStringToSDK(loadBalancers.Internal.SourceIPRanges),
			EndpointServiceAllowedPrincipals: common.ListStringToSDK(loadBalancers.Internal.EndpointServiceAllowedPrincipals),
			EndpointServiceSupportedRegions:  common.ListStringToSDK(loadBalancers.Internal.EndpointServiceSupportedRegions),
		}
	}

	return &sdk.HostedAWSEnvLoadBalancersSpecInput{
		Public:   public,
		Internal: internal,
	}
}

func loadBalancersToModel(loadBalancers sdk.HostedAWSEnvSpecFragment_LoadBalancers) *LoadBalancersModel {
	return &LoadBalancersModel{
		Public: &PublicLoadBalancerModel{
			Enabled:        types.BoolValue(loadBalancers.Public.Enabled),
			SourceIPRanges: common.ListStringToModel(loadBalancers.Public.SourceIPRanges),
		},
		Internal: &InternalLoadBalancerModel{
			Enabled:                          types.BoolValue(loadBalancers.Internal.Enabled),
			SourceIPRanges:                   common.ListStringToModel(loadBalancers.Internal.SourceIPRanges),
			EndpointServiceAllowedPrincipals: common.ListStringToModel(loadBalancers.Internal.EndpointServiceAllowedPrincipals),
			EndpointServiceSupportedRegions:  common.ListStringToModel(loadBalancers.Internal.EndpointServiceSupportedRegions),
		},
	}
}

func nodeGroupsToSDK(ctx context.Context, nodeGroups []hosted.NodeGroupsModel) ([]*sdk.HostedAWSEnvNodeGroupSpecInput, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var sdkNodeGroups []*sdk.HostedAWSEnvNodeGroupSpecInput
	for _, ng := range nodeGroups {
		var reservations []sdk.NodeReservation
		if !ng.Reservations.IsUnknown() && !ng.Reservations.IsNull() {
			allDiags.Append(ng.Reservations.ElementsAs(ctx, &reservations, false)...)
		}

		var zoneIDs []string
		if !ng.ZoneIDs.IsUnknown() && !ng.ZoneIDs.IsNull() {
			allDiags.Append(ng.ZoneIDs.ElementsAs(ctx, &zoneIDs, false)...)
		}

		sdkNodeGroups = append(sdkNodeGroups, &sdk.HostedAWSEnvNodeGroupSpecInput{
			Name:            ng.Name.ValueStringPointer(),
			NodeType:        ng.NodeType.ValueString(),
			ZoneIDs:         zoneIDs,
			Reservations:    reservations,
			CapacityPerZone: ng.CapacityPerZone.ValueInt64(),
		})
	}

	return sdkNodeGroups, allDiags
}

func nodeGroupsToModel(nodeGroups []*sdk.HostedAWSEnvSpecFragment_NodeGroups) ([]hosted.NodeGroupsModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	var models []hosted.NodeGroupsModel
	for _, ng := range nodeGroups {
		zoneIDs, diags := common.ListToModel(ng.ZoneIDs)
		allDiags.Append(diags...)
		reservations, diags := common.ReservationsToModel(ng.Reservations)
		allDiags.Append(diags...)

		models = append(models, hosted.NodeGroupsModel{
			Name:            types.StringValue(ng.Name),
			NodeType:        types.StringValue(ng.NodeType),
			ZoneIDs:         zoneIDs,
			Reservations:    reservations,
			CapacityPerZone: types.Int64Value(ng.CapacityPerZone),
		})
	}

	return models, allDiags
}

func maintenanceWindowsToModel(input []*sdk.HostedAWSEnvSpecFragment_MaintenanceWindows) []common.MaintenanceWindowModel {
	var maintenanceWindows []common.MaintenanceWindowModel
	for _, mw := range input {
		var days []types.String
		for _, day := range mw.Days {
			days = append(days, types.StringValue(string(day)))
		}

		maintenanceWindows = append(maintenanceWindows, common.MaintenanceWindowModel{
			Name:          types.StringValue(mw.Name),
			Enabled:       types.BoolValue(mw.Enabled),
			Hour:          types.Int64Value(mw.Hour),
			LengthInHours: types.Int64Value(mw.LengthInHours),
			Days:          days,
		})
	}

	return maintenanceWindows
}

func endpointsToModel(input []*sdk.HostedAWSEnvSpecFragment_Endpoints) []EndpointModel {
	var endpoints []EndpointModel
	for _, e := range input {
		endpoints = append(endpoints, EndpointModel{
			ServiceName: types.StringValue(e.ServiceName),
			Alias:       types.StringPointerValue(e.Alias),
		})
	}

	return endpoints
}

func externalBucketsToModel(input []*sdk.HostedAWSEnvSpecFragment_ExternalBuckets) []ExternalBucketModel {
	var buckets []ExternalBucketModel
	for _, b := range input {
		buckets = append(buckets, ExternalBucketModel{
			Name:      types.StringValue(b.Name),
			KmsKeyArn: types.StringPointerValue(b.KmsKeyArn),
		})
	}

	return buckets
}

func backupsToSDK(backups *BackupsModel) *sdk.HostedAWSEnvBackupsSpecInput {
	if backups == nil || backups.CustomBucket == nil {
		return nil
	}

	return &sdk.HostedAWSEnvBackupsSpecInput{
		CustomBucket: &sdk.HostedAWSEnvBackupsCustomBucketSpecInput{
			Name:    backups.CustomBucket.Name.ValueString(),
			Region:  backups.CustomBucket.Region.ValueString(),
			RoleArn: backups.CustomBucket.RoleArn.ValueString(),
		},
	}
}

func backupsToModel(backups *sdk.HostedAWSEnvSpecFragment_Backups) *BackupsModel {
	if backups == nil || backups.CustomBucket == nil {
		return nil
	}

	return &BackupsModel{
		CustomBucket: &CustomBucketModel{
			Name:    types.StringValue(backups.CustomBucket.Name),
			Region:  types.StringValue(backups.CustomBucket.Region),
			RoleArn: types.StringValue(backups.CustomBucket.RoleArn),
		},
	}
}

func icebergCatalogsToSDK(iceberg *IcebergModel) []*sdk.HostedAWSEnvIcebergCatalogInputSpec {
	if iceberg == nil {
		return nil
	}

	var catalogs []*sdk.HostedAWSEnvIcebergCatalogInputSpec
	for _, c := range iceberg.Catalogs {
		catalog := &sdk.HostedAWSEnvIcebergCatalogInputSpec{
			Name:                   c.Name.ValueStringPointer(),
			Type:                   sdk.HostedAWSEnvIcebergCatalogTypeSpec(c.Type.ValueString()),
			CustomS3Bucket:         c.CustomS3Bucket.ValueStringPointer(),
			CustomS3BucketPath:     c.CustomS3BucketPath.ValueStringPointer(),
			CustomS3TableBucketArn: c.CustomS3TableBucketArn.ValueStringPointer(),
			Region:                 c.Region.ValueStringPointer(),
			AnonymousAccessEnabled: c.AnonymousAccessEnabled.ValueBoolPointer(),
		}

		if c.Maintenance != nil {
			catalog.Maintenance = &sdk.HostedAWSEnvIcebergCatalogMaintenanceInputSpec{
				Enabled: c.Maintenance.Enabled.ValueBool(),
			}
		}

		var watches []*sdk.HostedAWSEnvIcebergCatalogWatchInputSpec
		for _, w := range c.Watches {
			watches = append(watches, &sdk.HostedAWSEnvIcebergCatalogWatchInputSpec{
				Table:                        w.Table.ValueString(),
				PathsRelativeToTableLocation: common.ListStringToSDK(w.PathsRelativeToTableLocation),
			})
		}
		catalog.Watches = watches

		catalogs = append(catalogs, catalog)
	}

	return catalogs
}

func icebergToModel(iceberg *sdk.HostedAWSEnvSpecFragment_Iceberg) *IcebergModel {
	if iceberg == nil || len(iceberg.Catalogs) == 0 {
		return nil
	}

	var catalogs []IcebergCatalogModel
	for _, c := range iceberg.Catalogs {
		catalog := IcebergCatalogModel{
			Name:                   types.StringPointerValue(c.Name),
			Type:                   types.StringValue(string(c.Type)),
			CustomS3Bucket:         types.StringPointerValue(c.CustomS3Bucket),
			CustomS3BucketPath:     types.StringPointerValue(c.CustomS3BucketPath),
			CustomS3TableBucketArn: types.StringPointerValue(c.CustomS3TableBucketArn),
			Region:                 types.StringPointerValue(c.Region),
			AnonymousAccessEnabled: types.BoolPointerValue(c.AnonymousAccessEnabled),
			Maintenance: &IcebergCatalogMaintenanceModel{
				Enabled: types.BoolValue(c.Maintenance.Enabled),
			},
		}

		var watches []IcebergCatalogWatchModel
		for _, w := range c.Watches {
			watches = append(watches, IcebergCatalogWatchModel{
				Table:                        types.StringValue(w.Table),
				PathsRelativeToTableLocation: common.ListStringToModel(w.PathsRelativeToTableLocation),
			})
		}
		catalog.Watches = watches

		catalogs = append(catalogs, catalog)
	}

	return &IcebergModel{Catalogs: catalogs}
}

// name is optional, so two unnamed catalogs would collide on name alone; adding
// type narrows it, and ReorderByKey's positional fallback keeps it loss-safe.
func icebergCatalogKey(name, catalogType string) string {
	return name + "\x1f" + catalogType
}

func ptrString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// Reorders the API catalogs (and each catalog's watches) into the user's configured
// order to avoid drift. Mutates spec in place.
func reorderIceberg(model *IcebergModel, spec *sdk.HostedAWSEnvSpecFragment_Iceberg) {
	if model == nil || spec == nil {
		return
	}

	spec.Catalogs = common.ReorderByKey(model.Catalogs, spec.Catalogs,
		func(m IcebergCatalogModel) string {
			return icebergCatalogKey(m.Name.ValueString(), m.Type.ValueString())
		},
		func(s *sdk.HostedAWSEnvSpecFragment_Iceberg_Catalogs) string {
			return icebergCatalogKey(ptrString(s.Name), string(s.Type))
		},
	)

	for _, mc := range model.Catalogs {
		for _, sc := range spec.Catalogs {
			if icebergCatalogKey(mc.Name.ValueString(), mc.Type.ValueString()) != icebergCatalogKey(ptrString(sc.Name), string(sc.Type)) {
				continue
			}
			sc.Watches = common.ReorderByKey(mc.Watches, sc.Watches,
				func(m IcebergCatalogWatchModel) string { return m.Table.ValueString() },
				func(s *sdk.HostedAWSEnvSpecFragment_Iceberg_Catalogs_Watches) string { return s.Table },
			)
			break
		}
	}
}
