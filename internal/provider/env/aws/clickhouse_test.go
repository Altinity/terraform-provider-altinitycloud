package env

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func chList(t *testing.T, values ...string) types.List {
	t.Helper()
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return list
}

func minimalClusterModel(name string) common.ClickHouseClusterModel {
	return common.ClickHouseClusterModel{
		Name:         types.StringValue(name),
		Image:        types.StringNull(),
		InstanceType: types.StringValue("m6i.large"),
		Zones:        types.ListNull(types.StringType),
		Shards:       types.Int64Null(),
		Replicas:     types.Int64Null(),
		Stopped:      types.BoolValue(false),
		Keeper:       types.StringValue("keeper"),
		Disk:         &common.ClickHouseDiskModel{SizeGiB: types.Int64Value(100)},
	}
}

func TestClickHouseClustersToSDK(t *testing.T) {
	ctx := context.Background()

	t.Run("no clusters returns nil", func(t *testing.T) {
		got, diags := clickHouseClustersToSDK(ctx, nil)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("omitted optionals stay nil", func(t *testing.T) {
		got, diags := clickHouseClustersToSDK(ctx, []common.ClickHouseClusterModel{minimalClusterModel("ch")})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		c := got[0]
		if c.Image != nil || c.Shards != nil || c.Replicas != nil || c.Zones != nil {
			t.Errorf("omitted optionals must stay nil, got %#v", c)
		}
		if c.ExtraDisks != nil || c.Users != nil || c.Profiles != nil || c.Quotas != nil || c.ServerSettings != nil {
			t.Errorf("omitted collections must stay nil, got %#v", c)
		}
		if c.Disk.Iops != nil || c.Disk.ThroughputInMiBps != nil {
			t.Errorf("omitted disk attributes must stay nil, got %#v", c.Disk)
		}
	})

	t.Run("unknown attributes are omitted, never sent as empty values", func(t *testing.T) {
		model := minimalClusterModel("ch")
		model.Image = types.StringUnknown()
		model.Keeper = types.StringUnknown()
		model.Zones = types.ListUnknown(types.StringType)
		model.Disk = &common.ClickHouseDiskModel{
			SizeGiB:        types.Int64Value(100),
			Iops:           types.Int64Unknown(),
			ThroughputMiBs: types.Int64Unknown(),
		}

		got, diags := clickHouseClustersToSDK(ctx, []common.ClickHouseClusterModel{model})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		c := got[0]
		if c.Image != nil || c.Keeper != nil || c.Zones != nil {
			t.Errorf("unknown attributes must be omitted, got %#v", c)
		}
		if c.Disk.Iops != nil || c.Disk.ThroughputInMiBps != nil {
			t.Errorf("unknown disk attributes must be omitted, got %#v", c.Disk)
		}
	})

	t.Run("every field round trips", func(t *testing.T) {
		model := minimalClusterModel("ch")
		model.Image = types.StringValue("altinity/clickhouse-server:25.8")
		model.Zones = chList(t, "us-east-1a", "us-east-1b")
		model.Shards = types.Int64Value(2)
		model.Replicas = types.Int64Value(3)
		model.Disk = &common.ClickHouseDiskModel{
			SizeGiB:        types.Int64Value(500),
			Iops:           types.Int64Value(3000),
			ThroughputMiBs: types.Int64Value(125),
		}
		model.ExtraDisks = []common.ClickHouseNamedDiskModel{{
			Name:    types.StringValue("disk1"),
			SizeGiB: types.Int64Value(50),
		}}
		model.ServerSettings = []common.ClickHouseSettingModel{{
			Key:   types.StringValue("max_concurrent_queries"),
			Value: types.StringValue("200"),
		}}
		model.Profiles = []common.ClickHouseProfileModel{{
			Name:     types.StringValue("readonly"),
			Settings: []common.ClickHouseSettingModel{{Key: types.StringValue("readonly"), Value: types.StringValue("1")}},
		}}
		model.Quotas = []common.ClickHouseQuotaModel{{
			Name:    types.StringValue("q"),
			KeyedBy: types.StringValue("IP_ADDRESS"),
			Intervals: []common.ClickHouseQuotaIntervalModel{{
				DurationSeconds: types.Int64Value(3600),
				Queries:         types.Int64Value(1000),
				ReadRows:        types.Int64Null(),
			}},
		}}
		model.Users = []common.ClickHouseUserModel{{
			Name:      types.StringValue("app"),
			Profile:   types.StringValue("readonly"),
			Networks:  &common.ClickHouseUserNetworksModel{Ips: chList(t, "10.0.0.0/8")},
			Databases: chList(t, "analytics"),
			Grants:    chList(t, "SELECT ON analytics.*"),
			Password: &common.ClickHouseUserPasswordModel{
				Type:     types.StringValue("SHA256_HEX"),
				EncValue: types.StringValue("deadbeef"),
			},
		}}

		got, diags := clickHouseClustersToSDK(ctx, []common.ClickHouseClusterModel{model})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		c := got[0]
		if len(c.Zones) != 2 || *c.Shards != 2 || *c.Replicas != 3 {
			t.Errorf("unexpected topology: %#v", c)
		}
		if c.Disk.SizeInGiB != 500 || *c.Disk.Iops != 3000 || *c.Disk.ThroughputInMiBps != 125 {
			t.Errorf("unexpected disk: %#v", c.Disk)
		}
		if len(c.ExtraDisks) != 1 || c.ExtraDisks[0].Name != "disk1" || c.ExtraDisks[0].SizeInGiB != 50 {
			t.Errorf("unexpected extra disks: %#v", c.ExtraDisks)
		}
		if len(c.ServerSettings) != 1 || *c.ServerSettings[0].Value != "200" {
			t.Errorf("unexpected server settings: %#v", c.ServerSettings)
		}
		if len(c.Profiles) != 1 || len(c.Profiles[0].Settings) != 1 {
			t.Errorf("unexpected profiles: %#v", c.Profiles)
		}

		q := c.Quotas[0]
		if *q.KeyedBy != sdk.ClickHouseQuotaKeyedByIPAddress {
			t.Errorf("keyed_by = %v, want IP_ADDRESS", *q.KeyedBy)
		}
		if q.Intervals[0].Duration != 3600 || *q.Intervals[0].Queries != 1000 {
			t.Errorf("unexpected interval: %#v", q.Intervals[0])
		}
		if q.Intervals[0].ReadRows != nil {
			t.Errorf("an unset limit must stay nil, got %v", *q.Intervals[0].ReadRows)
		}

		u := c.Users[0]
		if u.Password.Type != sdk.AWSEnvClickHouseUserPasswordTypeSha256Hex || *u.Password.EncValue != "deadbeef" {
			t.Errorf("unexpected password: %#v", u.Password)
		}
		if len(u.Networks.Ips) != 1 || len(u.Databases) != 1 || u.Databases[0].Name != "analytics" {
			t.Errorf("unexpected user scoping: %#v", u)
		}
		if len(u.Grants) != 1 {
			t.Errorf("unexpected grants: %v", u.Grants)
		}
	})
}

func TestClickHouseKeepersToSDK(t *testing.T) {
	ctx := context.Background()

	t.Run("omitted optionals stay nil", func(t *testing.T) {
		got, diags := clickHouseKeepersToSDK(ctx, []common.ClickHouseKeeperModel{{
			Name:         types.StringValue("keeper"),
			InstanceType: types.StringValue("t4g.large"),
			Zones:        types.ListNull(types.StringType),
			Replicas:     types.Int64Null(),
			Stopped:      types.BoolNull(),
			Disk:         &common.ClickHouseDiskModel{SizeGiB: types.Int64Value(30)},
		}})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		k := got[0]
		if k.Zones != nil || k.Replicas != nil || k.Stopped != nil {
			t.Errorf("omitted optionals must stay nil, got %#v", k)
		}
		if k.Disk.SizeInGiB != 30 || k.Disk.Iops != nil {
			t.Errorf("unexpected disk: %#v", k.Disk)
		}
	})

	t.Run("every field round trips", func(t *testing.T) {
		got, diags := clickHouseKeepersToSDK(ctx, []common.ClickHouseKeeperModel{{
			Name:         types.StringValue("keeper"),
			InstanceType: types.StringValue("t4g.large"),
			Zones:        chList(t, "us-east-1a"),
			Replicas:     types.Int64Value(3),
			Stopped:      types.BoolValue(true),
			Disk:         &common.ClickHouseDiskModel{SizeGiB: types.Int64Value(30), Iops: types.Int64Value(3000)},
		}})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}

		k := got[0]
		if k.Name != "keeper" || *k.Replicas != 3 || !*k.Stopped || len(k.Zones) != 1 {
			t.Errorf("unexpected keeper: %#v", k)
		}
	})
}

func TestClickHouseClustersToSpec(t *testing.T) {
	specs := clickHouseClustersToSpec([]*sdk.AWSEnvSpecFragment_ClickHouseClusters{
		nil,
		{
			Name:         "ch",
			Image:        "altinity/clickhouse-server:25.8",
			InstanceType: "m6i.large",
			Zones:        []string{"us-east-1a"},
			Shards:       1,
			Replicas:     2,
			Keeper:       "keeper",
			Disk: sdk.AWSEnvSpecFragment_ClickHouseClusters_Disk{
				SizeInGiB: 100, Iops: 3000, ThroughputInMiBps: 125,
			},
			Users: []*sdk.AWSEnvSpecFragment_ClickHouseClusters_Users{{
				Name:     "app",
				Networks: sdk.AWSEnvSpecFragment_ClickHouseClusters_Users_Networks{Ips: []string{"10.0.0.0/8"}},
				Password: sdk.AWSEnvSpecFragment_ClickHouseClusters_Users_Password{
					Type: sdk.AWSEnvClickHouseUserPasswordTypeSha256Hex,
				},
				Databases: []*sdk.AWSEnvSpecFragment_ClickHouseClusters_Users_Databases{{Name: "analytics"}},
			}},
		},
	})

	// A nil fragment entry must not become a spec.
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}

	s := specs[0]
	if s.Keeper != "keeper" || s.Disk.SizeGiB != 100 || s.Disk.ThroughputMiBs != 125 {
		t.Errorf("unexpected cluster: %#v", s)
	}
	if len(s.Users) != 1 || s.Users[0].PasswordType != "SHA256_HEX" {
		t.Errorf("unexpected users: %#v", s.Users)
	}
	if len(s.Users[0].Databases) != 1 || s.Users[0].Databases[0] != "analytics" {
		t.Errorf("unexpected databases: %v", s.Users[0].Databases)
	}
	if len(s.Users[0].Ips) != 1 {
		t.Errorf("unexpected ips: %v", s.Users[0].Ips)
	}
}

func TestUpdateUsesPatchSoConsoleClustersSurvive(t *testing.T) {
	model := AWSEnvModel{
		Name:               types.StringValue("env"),
		Region:             types.StringValue("us-east-1"),
		CIDR:               types.StringValue("10.0.0.0/16"),
		AWSAccountID:       types.StringValue("111122223333"),
		Zones:              types.ListNull(types.StringType),
		ClickHouseClusters: []common.ClickHouseClusterModel{minimalClusterModel("mine")},
		ClickHouseKeepers: []common.ClickHouseKeeperModel{{
			Name:         types.StringValue("keeper"),
			InstanceType: types.StringValue("t4g.large"),
			Zones:        types.ListNull(types.StringType),
			Disk:         &common.ClickHouseDiskModel{SizeGiB: types.Int64Value(30)},
		}},
	}

	_, update, diags := model.toSDK(context.Background(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	// The flat list replaces everything the API exposes; the patch does not.
	if update.Spec.ClickHouseClusters != nil || update.Spec.ClickHouseKeepers != nil {
		t.Errorf("an update must not send the replace-all list: %#v", update.Spec)
	}
	if update.Spec.ClickHouseClustersPatch == nil || len(update.Spec.ClickHouseClustersPatch.Replace) != 1 {
		t.Fatalf("unexpected cluster patch: %#v", update.Spec.ClickHouseClustersPatch)
	}
	if update.Spec.ClickHouseClustersPatch.Replace[0].Name != "mine" {
		t.Errorf("unexpected cluster: %#v", update.Spec.ClickHouseClustersPatch.Replace[0])
	}
	if update.Spec.ClickHouseKeepersPatch == nil || len(update.Spec.ClickHouseKeepersPatch.Replace) != 1 {
		t.Errorf("unexpected keeper patch: %#v", update.Spec.ClickHouseKeepersPatch)
	}

	// Nothing was dropped from the configuration, so nothing is named for deletion.
	if update.Spec.ClickHouseClustersPatch.Delete != nil {
		t.Errorf("unexpected deletes: %v", update.Spec.ClickHouseClustersPatch.Delete)
	}
}

func TestUpdateNamesDroppedClickHouseEntriesForDeletion(t *testing.T) {
	prior := AWSEnvModel{
		ClickHouseClusters: []common.ClickHouseClusterModel{minimalClusterModel("keep"), minimalClusterModel("dropped")},
		ClickHouseKeepers:  []common.ClickHouseKeeperModel{{Name: types.StringValue("old-keeper")}},
	}
	planned := AWSEnvModel{ClickHouseClusters: []common.ClickHouseClusterModel{minimalClusterModel("keep")}}

	_, update, diags := planned.toSDK(context.Background(), &prior)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	clusters := update.Spec.ClickHouseClustersPatch
	if clusters == nil || len(clusters.Replace) != 1 || len(clusters.Delete) != 1 || clusters.Delete[0] != "dropped" {
		t.Errorf("expected keep replaced and dropped deleted, got %#v", clusters)
	}
	keepers := update.Spec.ClickHouseKeepersPatch
	if keepers == nil || keepers.Replace != nil || len(keepers.Delete) != 1 || keepers.Delete[0] != "old-keeper" {
		t.Errorf("expected only a delete for the dropped keeper, got %#v", keepers)
	}
}

func TestUpdateWithoutClickHouseSendsNoPatch(t *testing.T) {
	_, update, diags := AWSEnvModel{}.toSDK(context.Background(), &AWSEnvModel{})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if update.Spec.ClickHouseClustersPatch != nil || update.Spec.ClickHouseKeepersPatch != nil {
		t.Errorf("an env without clusters must leave the API's entries untouched, got %#v / %#v", update.Spec.ClickHouseClustersPatch, update.Spec.ClickHouseKeepersPatch)
	}
}

func TestClickHouseNamesToDelete(t *testing.T) {
	prior := []common.ClickHouseClusterModel{minimalClusterModel("keep"), minimalClusterModel("dropped")}
	planned := []common.ClickHouseClusterModel{minimalClusterModel("keep"), minimalClusterModel("new")}

	got := common.ClickHouseClusterNamesToDelete(prior, planned)
	if len(got) != 1 || got[0] != "dropped" {
		t.Errorf("expected [dropped], got %v", got)
	}

	// A console cluster is in neither list, so it is never named for deletion.
	if got := common.ClickHouseClusterNamesToDelete(nil, planned); got != nil {
		t.Errorf("expected nil with no prior state, got %v", got)
	}
	if got := common.ClickHouseClusterNamesToDelete(prior, prior); got != nil {
		t.Errorf("expected nil when nothing was dropped, got %v", got)
	}

	keepersPrior := []common.ClickHouseKeeperModel{{Name: types.StringValue("a")}, {Name: types.StringValue("b")}}
	keepersPlanned := []common.ClickHouseKeeperModel{{Name: types.StringValue("b")}}
	if got := common.ClickHouseKeeperNamesToDelete(keepersPrior, keepersPlanned); len(got) != 1 || got[0] != "a" {
		t.Errorf("expected [a], got %v", got)
	}
}

func TestClickHouseCollisionDetection(t *testing.T) {
	state := []common.ClickHouseClusterModel{minimalClusterModel("mine")}
	plan := []common.ClickHouseClusterModel{minimalClusterModel("mine"), minimalClusterModel("prod-ch")}

	added := common.ClickHouseClusterNamesAdded(state, plan)
	if len(added) != 1 || added[0] != "prod-ch" {
		t.Fatalf("expected [prod-ch], got %v", added)
	}

	// The console cluster carries the planned name: adding it would overwrite, not create.
	if got := common.ClickHouseCollisions(added, []string{"mine", "prod-ch"}); len(got) != 1 || got[0] != "prod-ch" {
		t.Errorf("expected [prod-ch], got %v", got)
	}
	if got := common.ClickHouseCollisions(added, []string{"mine"}); got != nil {
		t.Errorf("expected no collisions, got %v", got)
	}
	if got := common.ClickHouseCollisions(nil, []string{"mine"}); got != nil {
		t.Errorf("expected no collisions with nothing added, got %v", got)
	}
}

func TestRejectClickHouseCollisionsAtApply(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"awsEnv":{"name":"env","spec":{"clickHouseClusters":[{"name":"mine"},{"name":"prod-ch"}],"clickHouseKeepers":[{"name":"console-keeper"}]}}}}`)
	}))
	t.Cleanup(srv.Close)
	r := newAWSResource(sdk.NewClient(srv.Client(), srv.URL, nil))

	state := []common.ClickHouseClusterModel{minimalClusterModel("mine")}

	// A name that was unknown at plan time resolves to the console cluster's name by apply.
	diags := r.rejectClickHouseCollisions(context.Background(), "env",
		state, []common.ClickHouseClusterModel{minimalClusterModel("mine"), minimalClusterModel("prod-ch")},
		nil, []common.ClickHouseKeeperModel{{Name: types.StringValue("console-keeper")}})
	if len(diags.Errors()) != 2 {
		t.Fatalf("expected a cluster and a keeper collision, got %v", diags)
	}

	if diags := r.rejectClickHouseCollisions(context.Background(), "env",
		state, []common.ClickHouseClusterModel{minimalClusterModel("mine"), minimalClusterModel("fresh")},
		nil, nil); diags.HasError() {
		t.Fatalf("expected no collision for a new name, got %v", diags)
	}
}
