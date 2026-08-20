package hosted_env

import (
	"context"
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	hosted "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env_hosted/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringList(values ...string) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	list, _ := types.ListValue(types.StringType, elements)
	return list
}

func stringSet(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	set, _ := types.SetValue(types.StringType, elements)
	return set
}

func minimalModel() AWSEnvHostedResourceModel {
	return AWSEnvHostedResourceModel{
		Name:          types.StringValue("dummy-env"),
		Region:        types.StringValue("us-east-1"),
		ZoneIDs:       stringList("use1-az1", "use1-az2"),
		CustomDomains: types.ListNull(types.StringType),
		NodeGroups: []hosted.NodeGroupsModel{{
			NodeType:        types.StringValue("m6i.large"),
			CapacityPerZone: types.Int64Value(1),
			ZoneIDs:         types.ListNull(types.StringType),
			Reservations:    stringSet("SYSTEM"),
			Name:            types.StringNull(),
		}},
	}
}

func minimalSpec() *sdk.AWSEnvHostedSpecFragment {
	return &sdk.AWSEnvHostedSpecFragment{
		Region:         "us-east-1",
		Cidr:           "10.0.0.0/16",
		ZoneIDs:        []string{"use1-az1", "use1-az2"},
		ResourcePrefix: "altinity",
		CustomDomains:  []string{},
		NodeGroups: []*sdk.AWSEnvHostedSpecFragment_NodeGroups{{
			Name:            "m6i.large",
			NodeType:        "m6i.large",
			CapacityPerZone: 1,
			ZoneIDs:         []string{"use1-az1", "use1-az2"},
			Reservations:    []sdk.NodeReservation{sdk.NodeReservationSystem},
		}},
		MaintenanceWindows: []*sdk.AWSEnvHostedSpecFragment_MaintenanceWindows{},
		Endpoints:          []*sdk.AWSEnvHostedSpecFragment_Endpoints{},
		ExternalBuckets:    []*sdk.AWSEnvHostedSpecFragment_ExternalBuckets{},
	}
}

func TestToSDK_MinimalConfigLeavesOptionalInputsUnset(t *testing.T) {
	model := minimalModel()

	create, update, diags := model.toSDK(context.Background())
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if create.Spec.Region != "us-east-1" {
		t.Errorf("region = %q, want us-east-1", create.Spec.Region)
	}
	if got := create.Spec.ZoneIDs; len(got) != 2 {
		t.Errorf("zone_ids = %v, want 2 entries", got)
	}
	if create.Spec.CustomDomains != nil {
		t.Errorf("custom_domains = %v, want nil", create.Spec.CustomDomains)
	}
	if create.Spec.LoadBalancers != nil {
		t.Errorf("load_balancers = %v, want nil", create.Spec.LoadBalancers)
	}
	if create.Spec.Backups != nil {
		t.Errorf("backups = %v, want nil", create.Spec.Backups)
	}
	if create.Spec.Iceberg != nil {
		t.Errorf("iceberg = %v, want nil", create.Spec.Iceberg)
	}
	if create.Spec.MetricsEndpoint != nil {
		t.Errorf("metrics_endpoint = %v, want nil", create.Spec.MetricsEndpoint)
	}
	if create.Spec.Datadog != nil {
		t.Errorf("datadog = %v, want nil", create.Spec.Datadog)
	}
	if create.Spec.NodeGroups[0].ZoneIDs != nil {
		t.Errorf("node group zone_ids = %v, want nil", create.Spec.NodeGroups[0].ZoneIDs)
	}

	if update.UpdateStrategy == nil || *update.UpdateStrategy != sdk.UpdateStrategyReplace {
		t.Errorf("update strategy = %v, want REPLACE", update.UpdateStrategy)
	}
	if update.Spec.Iceberg != nil {
		t.Errorf("update iceberg = %v, want nil", update.Spec.Iceberg)
	}
}

func TestApplySpec_MinimalConfigKeepsOmittedAttributesNull(t *testing.T) {
	model := minimalModel()

	if diags := model.applySpec(context.Background(), "dummy-env", minimalSpec(), 7); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !model.CustomDomains.IsNull() {
		t.Errorf("custom_domains = %v, want null", model.CustomDomains)
	}
	if model.Endpoints != nil {
		t.Errorf("endpoints = %v, want nil", model.Endpoints)
	}
	if model.ExternalBuckets != nil {
		t.Errorf("external_buckets = %v, want nil", model.ExternalBuckets)
	}
	if model.MaintenanceWindows != nil {
		t.Errorf("maintenance_windows = %v, want nil", model.MaintenanceWindows)
	}
	if model.Backups != nil || model.Iceberg != nil {
		t.Errorf("backups/iceberg = %v/%v, want nil", model.Backups, model.Iceberg)
	}
	if model.MetricsEndpoint != nil {
		t.Errorf("metrics_endpoint = %v, want nil when disabled and unconfigured", model.MetricsEndpoint)
	}
	if model.Datadog != nil {
		t.Errorf("datadog = %v, want nil when disabled and unconfigured", model.Datadog)
	}
	if model.CIDR.ValueString() != "10.0.0.0/16" {
		t.Errorf("cidr = %q, want 10.0.0.0/16", model.CIDR.ValueString())
	}
	if model.SpecRevision.ValueInt64() != 7 {
		t.Errorf("spec_revision = %d, want 7", model.SpecRevision.ValueInt64())
	}
	if model.KmsKeyArn.IsNull() != true {
		t.Errorf("kms_key_arn = %v, want null", model.KmsKeyArn)
	}
}

func TestApplySpec_KeepsCustomDomainsWhenConfigured(t *testing.T) {
	model := minimalModel()
	model.CustomDomains = stringList("example.com")

	spec := minimalSpec()
	spec.CustomDomains = []string{"example.com"}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if model.CustomDomains.IsNull() || len(model.CustomDomains.Elements()) != 1 {
		t.Errorf("custom_domains = %v, want [example.com]", model.CustomDomains)
	}
}

func TestApplySpec_ReordersNodeGroupsAndZonesToConfigOrder(t *testing.T) {
	model := minimalModel()
	model.ZoneIDs = stringList("use1-az2", "use1-az1")
	model.NodeGroups = []hosted.NodeGroupsModel{
		{NodeType: types.StringValue("m6i.xlarge"), ZoneIDs: stringList("use1-az2", "use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(1)},
		{NodeType: types.StringValue("m6i.large"), ZoneIDs: stringList("use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(1)},
	}

	spec := minimalSpec()
	spec.NodeGroups = []*sdk.AWSEnvHostedSpecFragment_NodeGroups{
		{Name: "m6i.large", NodeType: "m6i.large", CapacityPerZone: 1, ZoneIDs: []string{"use1-az1"}},
		{Name: "m6i.xlarge", NodeType: "m6i.xlarge", CapacityPerZone: 1, ZoneIDs: []string{"use1-az1", "use1-az2"}},
	}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.NodeGroups[0].NodeType.ValueString(); got != "m6i.xlarge" {
		t.Errorf("node_groups[0].node_type = %q, want m6i.xlarge", got)
	}
	if got := firstString(t, model.NodeGroups[0].ZoneIDs); got != "use1-az2" {
		t.Errorf("node_groups[0].zone_ids[0] = %q, want use1-az2", got)
	}
	if got := firstString(t, model.ZoneIDs); got != "use1-az2" {
		t.Errorf("zone_ids[0] = %q, want use1-az2", got)
	}
}

func TestApplySpec_ReordersNodeGroupsByNameWhenNodeTypesRepeat(t *testing.T) {
	model := minimalModel()
	model.NodeGroups = []hosted.NodeGroupsModel{
		{Name: types.StringValue("workers-b"), NodeType: types.StringValue("m6i.large"), ZoneIDs: stringList("use1-az2", "use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(2)},
		{Name: types.StringValue("workers-a"), NodeType: types.StringValue("m6i.large"), ZoneIDs: stringList("use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(1)},
	}

	spec := minimalSpec()
	spec.NodeGroups = []*sdk.AWSEnvHostedSpecFragment_NodeGroups{
		{Name: "workers-a", NodeType: "m6i.large", CapacityPerZone: 1, ZoneIDs: []string{"use1-az1"}},
		{Name: "workers-b", NodeType: "m6i.large", CapacityPerZone: 2, ZoneIDs: []string{"use1-az1", "use1-az2"}},
	}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.NodeGroups[0].Name.ValueString(); got != "workers-b" {
		t.Errorf("node_groups[0].name = %q, want workers-b", got)
	}
	if got := model.NodeGroups[0].CapacityPerZone.ValueInt64(); got != 2 {
		t.Errorf("node_groups[0].capacity_per_zone = %d, want 2", got)
	}
	if got := firstString(t, model.NodeGroups[0].ZoneIDs); got != "use1-az2" {
		t.Errorf("node_groups[0].zone_ids[0] = %q, want use1-az2", got)
	}
	if got := model.NodeGroups[1].Name.ValueString(); got != "workers-a" {
		t.Errorf("node_groups[1].name = %q, want workers-a", got)
	}
}

func TestApplySpec_PairsUnknownNameNodeGroupsByNodeType(t *testing.T) {
	model := minimalModel()
	model.NodeGroups = []hosted.NodeGroupsModel{
		{Name: types.StringUnknown(), NodeType: types.StringValue("m6i.xlarge"), ZoneIDs: stringList("use1-az2", "use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(1)},
		{Name: types.StringUnknown(), NodeType: types.StringValue("m6i.large"), ZoneIDs: stringList("use1-az1"), Reservations: stringSet("SYSTEM"), CapacityPerZone: types.Int64Value(1)},
	}

	spec := minimalSpec()
	spec.NodeGroups = []*sdk.AWSEnvHostedSpecFragment_NodeGroups{
		{Name: "m6i.large", NodeType: "m6i.large", CapacityPerZone: 1, ZoneIDs: []string{"use1-az1"}},
		{Name: "m6i.xlarge", NodeType: "m6i.xlarge", CapacityPerZone: 1, ZoneIDs: []string{"use1-az1", "use1-az2"}},
	}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.NodeGroups[0].NodeType.ValueString(); got != "m6i.xlarge" {
		t.Errorf("node_groups[0].node_type = %q, want m6i.xlarge", got)
	}
	if got := firstString(t, model.NodeGroups[0].ZoneIDs); got != "use1-az2" {
		t.Errorf("node_groups[0].zone_ids[0] = %q, want use1-az2", got)
	}
}

func TestApplySpec_ReordersIcebergCatalogsAndWatches(t *testing.T) {
	model := minimalModel()
	model.Iceberg = &IcebergModel{Catalogs: []IcebergCatalogModel{
		{Name: types.StringValue("b"), Type: types.StringValue("S3")},
		{
			Name: types.StringValue("a"), Type: types.StringValue("S3"),
			Watches: []IcebergCatalogWatchModel{
				{Table: types.StringValue("t2")},
				{Table: types.StringValue("t1")},
			},
		},
	}}

	catalogA := "a"
	catalogB := "b"
	spec := minimalSpec()
	spec.Iceberg = &sdk.AWSEnvHostedSpecFragment_Iceberg{
		Catalogs: []*sdk.AWSEnvHostedSpecFragment_Iceberg_Catalogs{
			{Name: &catalogA, Type: sdk.AWSEnvHostedIcebergCatalogTypeSpecS3, Watches: []*sdk.AWSEnvHostedSpecFragment_Iceberg_Catalogs_Watches{
				{Table: "t1"},
				{Table: "t2"},
			}},
			{Name: &catalogB, Type: sdk.AWSEnvHostedIcebergCatalogTypeSpecS3},
		},
	}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.Iceberg.Catalogs[0].Name.ValueString(); got != "b" {
		t.Errorf("catalogs[0].name = %q, want b", got)
	}
	if got := model.Iceberg.Catalogs[1].Watches[0].Table.ValueString(); got != "t2" {
		t.Errorf("catalogs[1].watches[0].table = %q, want t2", got)
	}
}

func TestApplySpec_MissingSpecIsAnError(t *testing.T) {
	model := minimalModel()

	if diags := model.applySpec(context.Background(), "dummy-env", nil, 1); !diags.HasError() {
		t.Fatal("expected an error diagnostic for a nil spec")
	}
}

func TestApplySpec_PreservesDatadogEncAPIKey(t *testing.T) {
	model := minimalModel()
	model.Datadog = &common.DatadogModel{
		Enabled:   types.BoolValue(true),
		EncAPIKey: types.StringValue("secret"),
	}

	spec := minimalSpec()
	spec.Datadog = sdk.AWSEnvHostedSpecFragment_Datadog{Enabled: true, Domain: "datadoghq.com"}

	if diags := model.applySpec(context.Background(), "dummy-env", spec, 1); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := model.Datadog.EncAPIKey.ValueString(); got != "secret" {
		t.Errorf("datadog.enc_api_key = %q, want secret (write-only, never returned)", got)
	}
}

func firstString(t *testing.T, list types.List) string {
	t.Helper()

	value, ok := list.Elements()[0].(types.String)
	if !ok {
		t.Fatalf("element 0 of %v is not a string", list)
	}
	return value.ValueString()
}

func TestLoadBalancersToSDK(t *testing.T) {
	t.Run("nil model yields nil input", func(t *testing.T) {
		if got := loadBalancersToSDK(nil); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("only the configured side is sent", func(t *testing.T) {
		got := loadBalancersToSDK(&LoadBalancersModel{
			Public: &PublicLoadBalancerModel{
				Enabled:        types.BoolValue(true),
				SourceIPRanges: []types.String{types.StringValue("0.0.0.0/0")},
			},
		})
		if got.Internal != nil {
			t.Errorf("internal = %v, want nil", got.Internal)
		}
		if got.Public.Enabled == nil || !*got.Public.Enabled {
			t.Errorf("public.enabled = %v, want true", got.Public.Enabled)
		}
	})

	t.Run("internal carries the endpoint service fields", func(t *testing.T) {
		got := loadBalancersToSDK(&LoadBalancersModel{
			Internal: &InternalLoadBalancerModel{
				Enabled:                          types.BoolValue(true),
				EndpointServiceAllowedPrincipals: []types.String{types.StringValue("arn:aws:iam::123456789012:root")},
				EndpointServiceSupportedRegions:  []types.String{types.StringValue("us-east-1")},
			},
		})
		if len(got.Internal.EndpointServiceAllowedPrincipals) != 1 {
			t.Errorf("allowed_principals = %v, want 1 entry", got.Internal.EndpointServiceAllowedPrincipals)
		}
		if len(got.Internal.EndpointServiceSupportedRegions) != 1 {
			t.Errorf("supported_regions = %v, want 1 entry", got.Internal.EndpointServiceSupportedRegions)
		}
	})
}

func TestLoadBalancersToModel(t *testing.T) {
	got := loadBalancersToModel(sdk.AWSEnvHostedSpecFragment_LoadBalancers{
		Public: sdk.AWSEnvHostedSpecFragment_LoadBalancers_Public{
			Enabled:        true,
			SourceIPRanges: []string{"0.0.0.0/0"},
		},
		Internal: sdk.AWSEnvHostedSpecFragment_LoadBalancers_Internal{
			Enabled:                         false,
			EndpointServiceSupportedRegions: []string{"us-east-1"},
		},
	})

	if !got.Public.Enabled.ValueBool() || len(got.Public.SourceIPRanges) != 1 {
		t.Errorf("public = %+v, want enabled with one range", got.Public)
	}
	if got.Internal.Enabled.ValueBool() {
		t.Errorf("internal.enabled = true, want false")
	}
	// nil slice keeps the attribute null instead of turning into an empty list.
	if got.Internal.SourceIPRanges != nil {
		t.Errorf("internal.source_ip_ranges = %v, want nil", got.Internal.SourceIPRanges)
	}
}

func TestNodeGroupsToSDK(t *testing.T) {
	got, diags := nodeGroupsToSDK(context.Background(), []hosted.NodeGroupsModel{
		{
			Name:            types.StringValue("system"),
			NodeType:        types.StringValue("m6i.large"),
			CapacityPerZone: types.Int64Value(3),
			ZoneIDs:         stringList("use1-az1"),
			Reservations:    stringSet("SYSTEM", "CLICKHOUSE"),
		},
		{
			NodeType:        types.StringValue("t4g.large"),
			CapacityPerZone: types.Int64Value(1),
			ZoneIDs:         types.ListNull(types.StringType),
			Reservations:    stringSet("ZOOKEEPER"),
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got[0].Name == nil || *got[0].Name != "system" {
		t.Errorf("node_groups[0].name = %v, want system", got[0].Name)
	}
	if got[0].CapacityPerZone != 3 || len(got[0].ZoneIDs) != 1 {
		t.Errorf("node_groups[0] = %+v, want capacity 3 and one zone id", got[0])
	}
	if len(got[0].Reservations) != 2 {
		t.Errorf("node_groups[0].reservations = %v, want 2 entries", got[0].Reservations)
	}
	if got[1].Name != nil {
		t.Errorf("node_groups[1].name = %v, want nil so the API defaults it", got[1].Name)
	}
	if got[1].ZoneIDs != nil {
		t.Errorf("node_groups[1].zone_ids = %v, want nil so the API defaults to the env zones", got[1].ZoneIDs)
	}
}

func TestNodeGroupsToModel(t *testing.T) {
	got, diags := nodeGroupsToModel([]*sdk.AWSEnvHostedSpecFragment_NodeGroups{{
		Name:            "m6i.large",
		NodeType:        "m6i.large",
		CapacityPerZone: 2,
		ZoneIDs:         []string{"use1-az1", "use1-az2"},
		Reservations:    []sdk.NodeReservation{sdk.NodeReservationSystem},
	}})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got[0].Name.ValueString() != "m6i.large" || got[0].CapacityPerZone.ValueInt64() != 2 {
		t.Errorf("got %+v, want the server-assigned name and capacity", got[0])
	}
	if len(got[0].ZoneIDs.Elements()) != 2 || len(got[0].Reservations.Elements()) != 1 {
		t.Errorf("got %+v, want 2 zone ids and 1 reservation", got[0])
	}
}

func TestBackupsToSDK(t *testing.T) {
	if got := backupsToSDK(nil); got != nil {
		t.Errorf("nil model: got %v, want nil", got)
	}
	if got := backupsToSDK(&BackupsModel{}); got != nil {
		t.Errorf("model without custom bucket: got %v, want nil", got)
	}

	got := backupsToSDK(&BackupsModel{CustomBucket: &CustomBucketModel{
		Name:    types.StringValue("bucket"),
		Region:  types.StringValue("us-east-1"),
		RoleArn: types.StringValue("arn:aws:iam::123456789012:role/backup"),
	}})
	if got.CustomBucket.Name != "bucket" || got.CustomBucket.RoleArn == "" {
		t.Errorf("got %+v, want the bucket fields carried over", got.CustomBucket)
	}
}

func TestBackupsToModel(t *testing.T) {
	if got := backupsToModel(nil); got != nil {
		t.Errorf("nil spec: got %v, want nil", got)
	}
	if got := backupsToModel(&sdk.AWSEnvHostedSpecFragment_Backups{}); got != nil {
		t.Errorf("spec without custom bucket: got %v, want nil", got)
	}

	got := backupsToModel(&sdk.AWSEnvHostedSpecFragment_Backups{
		CustomBucket: &sdk.AWSEnvHostedSpecFragment_Backups_CustomBucket{
			Name: "bucket", Region: "us-east-1", RoleArn: "arn:aws:iam::123456789012:role/backup",
		},
	})
	if got.CustomBucket.Name.ValueString() != "bucket" {
		t.Errorf("got %+v, want name bucket", got.CustomBucket)
	}
}

func TestIcebergCatalogsToSDK(t *testing.T) {
	if got := icebergCatalogsToSDK(nil); got != nil {
		t.Errorf("nil model: got %v, want nil", got)
	}

	got := icebergCatalogsToSDK(&IcebergModel{Catalogs: []IcebergCatalogModel{{
		Name:                   types.StringValue("analytics"),
		Type:                   types.StringValue("S3"),
		CustomS3Bucket:         types.StringValue("acme-iceberg"),
		CustomS3BucketPath:     types.StringValue("warehouse"),
		Region:                 types.StringValue("us-east-1"),
		AnonymousAccessEnabled: types.BoolValue(true),
		Maintenance:            &IcebergCatalogMaintenanceModel{Enabled: types.BoolValue(false)},
		Watches: []IcebergCatalogWatchModel{{
			Table:                        types.StringValue("events"),
			PathsRelativeToTableLocation: []types.String{types.StringValue("data")},
		}},
	}}})

	if got[0].Type != sdk.AWSEnvHostedIcebergCatalogTypeSpecS3 {
		t.Errorf("type = %v, want S3", got[0].Type)
	}
	if got[0].CustomS3TableBucketArn != nil {
		t.Errorf("custom_s3_table_bucket_arn = %v, want nil when unset", got[0].CustomS3TableBucketArn)
	}
	if got[0].Maintenance == nil || got[0].Maintenance.Enabled {
		t.Errorf("maintenance = %+v, want enabled=false", got[0].Maintenance)
	}
	if len(got[0].Watches) != 1 || got[0].Watches[0].Table != "events" {
		t.Errorf("watches = %+v, want one watch on events", got[0].Watches)
	}
}

func TestIcebergToModel(t *testing.T) {
	if got := icebergToModel(nil); got != nil {
		t.Errorf("nil spec: got %v, want nil", got)
	}
	// An env with iceberg disabled comes back with an empty catalog list; state must stay null.
	if got := icebergToModel(&sdk.AWSEnvHostedSpecFragment_Iceberg{}); got != nil {
		t.Errorf("empty catalogs: got %v, want nil", got)
	}

	name := "analytics"
	bucket := "acme-iceberg"
	got := icebergToModel(&sdk.AWSEnvHostedSpecFragment_Iceberg{
		Catalogs: []*sdk.AWSEnvHostedSpecFragment_Iceberg_Catalogs{{
			Name:           &name,
			Type:           sdk.AWSEnvHostedIcebergCatalogTypeSpecS3,
			CustomS3Bucket: &bucket,
			Maintenance:    sdk.AWSEnvHostedSpecFragment_Iceberg_Catalogs_Maintenance{Enabled: true},
			Watches: []*sdk.AWSEnvHostedSpecFragment_Iceberg_Catalogs_Watches{
				{Table: "events", PathsRelativeToTableLocation: []string{"data"}},
			},
		}},
	})

	catalog := got.Catalogs[0]
	if catalog.Name.ValueString() != "analytics" || catalog.CustomS3Bucket.ValueString() != "acme-iceberg" {
		t.Errorf("got %+v, want the catalog identity carried over", catalog)
	}
	if !catalog.Maintenance.Enabled.ValueBool() {
		t.Errorf("maintenance.enabled = false, want true")
	}
	if catalog.CustomS3TableBucketArn.IsNull() != true || catalog.Region.IsNull() != true {
		t.Errorf("unset optional fields should stay null, got %+v", catalog)
	}
	if len(catalog.Watches[0].PathsRelativeToTableLocation) != 1 {
		t.Errorf("watch paths = %v, want 1 entry", catalog.Watches[0].PathsRelativeToTableLocation)
	}
}

func TestReorderIcebergNilSafe(t *testing.T) {
	reorderIceberg(nil, nil)
	reorderIceberg(&IcebergModel{}, nil)
	reorderIceberg(nil, &sdk.AWSEnvHostedSpecFragment_Iceberg{})
}

func TestMaintenanceWindowsToModel(t *testing.T) {
	if got := common.MaintenanceWindowsToModel[*sdk.AWSEnvHostedSpecFragment_MaintenanceWindows](nil); got != nil {
		t.Errorf("got %v, want nil so an unconfigured attribute stays null", got)
	}

	got := common.MaintenanceWindowsToModel([]*sdk.AWSEnvHostedSpecFragment_MaintenanceWindows{{
		Name:          "weekly",
		Enabled:       true,
		Hour:          2,
		LengthInHours: 4,
		Days:          []sdk.Day{sdk.DayMonday, sdk.DayTuesday},
	}})

	if got[0].Name.ValueString() != "weekly" || got[0].Hour.ValueInt64() != 2 || len(got[0].Days) != 2 {
		t.Errorf("got %+v, want the window carried over with 2 days", got[0])
	}
}

func TestEndpointsToModel(t *testing.T) {
	if got := endpointsToModel(nil); got != nil {
		t.Errorf("got %v, want nil so an unconfigured attribute stays null", got)
	}

	alias := "kafka"
	got := endpointsToModel([]*sdk.AWSEnvHostedSpecFragment_Endpoints{
		{ServiceName: "com.amazonaws.vpce.us-east-1.vpce-svc-1", Alias: &alias},
		{ServiceName: "com.amazonaws.vpce.us-east-1.vpce-svc-2"},
	})

	if got[0].Alias.ValueString() != "kafka" {
		t.Errorf("endpoints[0].alias = %v, want kafka", got[0].Alias)
	}
	if !got[1].Alias.IsNull() {
		t.Errorf("endpoints[1].alias = %v, want null", got[1].Alias)
	}
}

func TestExternalBucketsToModel(t *testing.T) {
	if got := externalBucketsToModel(nil); got != nil {
		t.Errorf("got %v, want nil so an unconfigured attribute stays null", got)
	}

	arn := "arn:aws:kms:us-east-1:123456789012:key/abc"
	got := externalBucketsToModel([]*sdk.AWSEnvHostedSpecFragment_ExternalBuckets{
		{Name: "with-kms", KmsKeyArn: &arn},
		{Name: "plain"},
	})

	if got[0].KmsKeyArn.ValueString() != arn {
		t.Errorf("external_buckets[0].kms_key_arn = %v, want %s", got[0].KmsKeyArn, arn)
	}
	if !got[1].KmsKeyArn.IsNull() {
		t.Errorf("external_buckets[1].kms_key_arn = %v, want null", got[1].KmsKeyArn)
	}
}
