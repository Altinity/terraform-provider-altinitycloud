package env

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testList(t *testing.T, values ...string) types.List {
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

func listStrings(t *testing.T, list types.List) []string {
	t.Helper()
	var values []string
	for _, e := range list.Elements() {
		v, ok := e.(types.String)
		if !ok {
			t.Fatalf("expected a string element, got %T", e)
		}
		values = append(values, v.ValueString())
	}
	return values
}

func minimalClusterSpec(name string) ClickHouseClusterSpec {
	return ClickHouseClusterSpec{
		Name:         name,
		Image:        "altinity/clickhouse-server:25.8",
		InstanceType: "m6i.large",
		Zones:        []string{"us-east-1a"},
		Shards:       1,
		Replicas:     2,
		Keeper:       "keeper",
		Disk:         ClickHouseDiskSpec{SizeGiB: 100, Iops: 3000, ThroughputMiBs: 125},
	}
}

func TestClickHouseClustersToModel(t *testing.T) {
	t.Run("without prior state nothing is adopted", func(t *testing.T) {
		got := ClickHouseClustersToModel(nil, []ClickHouseClusterSpec{minimalClusterSpec("ch")})
		if got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("a cluster absent from prior state is left alone", func(t *testing.T) {
		prior := []ClickHouseClusterModel{{Name: types.StringValue("mine")}}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{
			minimalClusterSpec("mine"),
			minimalClusterSpec("made-in-the-console"),
		})

		if len(got) != 1 || got[0].Name.ValueString() != "mine" {
			t.Errorf("expected only the managed cluster, got %#v", got)
		}
	})

	t.Run("clusters come back in config order", func(t *testing.T) {
		prior := []ClickHouseClusterModel{
			{Name: types.StringValue("b")},
			{Name: types.StringValue("a")},
		}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{
			minimalClusterSpec("a"),
			minimalClusterSpec("b"),
		})

		if len(got) != 2 || got[0].Name.ValueString() != "b" || got[1].Name.ValueString() != "a" {
			t.Errorf("expected config order [b a], got %#v", got)
		}
	})

	t.Run("empty API lists stay null, never empty lists", func(t *testing.T) {
		prior := []ClickHouseClusterModel{{Name: types.StringValue("ch")}}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{minimalClusterSpec("ch")})

		c := got[0]
		if c.ExtraDisks != nil || c.Users != nil || c.Profiles != nil || c.Quotas != nil || c.ServerSettings != nil {
			t.Errorf("omitted collections must stay nil, got %#v", c)
		}
	})
}

func TestClickHouseUsersToModel(t *testing.T) {
	t.Run("reserved users are dropped", func(t *testing.T) {
		prior := []ClickHouseUserModel{{Name: types.StringValue("app")}}
		got := usersToModel(prior, []ClickHouseUserSpec{
			{Name: "app", PasswordType: "SHA256_HEX"},
			{Name: "grafana", PasswordType: "SHA256_HEX"},
			{Name: "datadog", PasswordType: "SHA256_HEX"},
		})

		if len(got) != 1 || got[0].Name.ValueString() != "app" {
			t.Errorf("expected only app, got %#v", got)
		}
	})

	t.Run("enc_value comes from prior state, never from the API", func(t *testing.T) {
		prior := []ClickHouseUserModel{{
			Name: types.StringValue("app"),
			Password: &ClickHouseUserPasswordModel{
				Type:     types.StringValue("SHA256_HEX"),
				EncValue: types.StringValue("kept-from-config"),
			},
		}}
		got := usersToModel(prior, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})

		if got[0].Password.EncValue.ValueString() != "kept-from-config" {
			t.Errorf("enc_value = %v, want the configured value", got[0].Password.EncValue)
		}
	})

	t.Run("a user absent from prior state has a null enc_value", func(t *testing.T) {
		got := usersToModel([]ClickHouseUserModel{}, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})

		if !got[0].Password.EncValue.IsNull() {
			t.Errorf("enc_value = %v, want null", got[0].Password.EncValue)
		}
	})

	t.Run("an omitted networks block stays nil", func(t *testing.T) {
		prior := []ClickHouseUserModel{{Name: types.StringValue("app")}}
		got := usersToModel(prior, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})

		if got[0].Networks != nil {
			t.Errorf("networks = %#v, want nil", got[0].Networks)
		}
		if !got[0].Databases.IsNull() || !got[0].Grants.IsNull() {
			t.Errorf("omitted lists must stay null, got %#v / %#v", got[0].Databases, got[0].Grants)
		}
	})

	t.Run("ips follow config order", func(t *testing.T) {
		prior := []ClickHouseUserModel{{
			Name:     types.StringValue("app"),
			Networks: &ClickHouseUserNetworksModel{Ips: testList(t, "10.0.0.0/8", "192.168.0.0/16")},
		}}
		got := usersToModel(prior, []ClickHouseUserSpec{{
			Name:         "app",
			Ips:          []string{"192.168.0.0/16", "10.0.0.0/8"},
			PasswordType: "SHA256_HEX",
		}})

		ips := listStrings(t, got[0].Networks.Ips)
		if len(ips) != 2 || ips[0] != "10.0.0.0/8" {
			t.Errorf("expected config order, got %v", ips)
		}
	})
}

func TestClickHouseQuotasToModel(t *testing.T) {
	queries := uint64(1000)
	prior := []ClickHouseQuotaModel{{Name: types.StringValue("q")}}

	got := quotasToModel(prior, []ClickHouseQuotaSpec{{
		Name:    "q",
		KeyedBy: "IP_ADDRESS",
		Intervals: []ClickHouseQuotaIntervalSpec{{
			DurationSeconds: 3600,
			Queries:         &queries,
		}},
	}})

	q := got[0]
	if q.KeyedBy.ValueString() != "IP_ADDRESS" {
		t.Errorf("keyed_by = %v, want IP_ADDRESS", q.KeyedBy)
	}
	if !q.IPv4PrefixBits.IsNull() || !q.IPv6PrefixBits.IsNull() {
		t.Errorf("omitted prefix bits must stay null, got %v / %v", q.IPv4PrefixBits, q.IPv6PrefixBits)
	}
	i := q.Intervals[0]
	if i.DurationSeconds.ValueInt64() != 3600 || i.Queries.ValueInt64() != 1000 {
		t.Errorf("unexpected interval: %#v", i)
	}
	if !i.ReadRows.IsNull() || !i.ExecutionTimeSeconds.IsNull() {
		t.Errorf("unset limits must stay null, got %v / %v", i.ReadRows, i.ExecutionTimeSeconds)
	}
}

func TestClickHouseKeepersToModel(t *testing.T) {
	t.Run("without prior state nothing is adopted", func(t *testing.T) {
		got := ClickHouseKeepersToModel(nil, []ClickHouseKeeperSpec{{Name: "keeper"}})
		if got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("zones follow config order", func(t *testing.T) {
		prior := []ClickHouseKeeperModel{{
			Name:  types.StringValue("keeper"),
			Zones: testList(t, "us-east-1b", "us-east-1a"),
		}}
		got := ClickHouseKeepersToModel(prior, []ClickHouseKeeperSpec{{
			Name:     "keeper",
			Zones:    []string{"us-east-1a", "us-east-1b"},
			Replicas: 3,
			Disk:     ClickHouseDiskSpec{SizeGiB: 30},
		}})

		zones := listStrings(t, got[0].Zones)
		if zones[0] != "us-east-1b" {
			t.Errorf("expected config order, got %v", zones)
		}
		if got[0].Replicas.ValueInt64() != 3 {
			t.Errorf("replicas = %v, want 3", got[0].Replicas)
		}
	})
}

func TestClickHouseDataSourceModelReportsEverything(t *testing.T) {
	specs := []ClickHouseClusterSpec{minimalClusterSpec("mine"), minimalClusterSpec("made-in-the-console")}
	specs[0].Users = []ClickHouseUserSpec{
		{Name: "app", PasswordType: "SHA256_HEX"},
		{Name: "grafana", PasswordType: "SHA256_HEX"},
	}

	// The resource mapper owns nothing without prior state; the data source
	// must still report the environment as it is.
	if got := ClickHouseClustersToModel(nil, specs); got != nil {
		t.Errorf("resource mapper adopted %#v", got)
	}

	got := ClickHouseClustersToDataSourceModel(specs)
	if len(got) != 2 {
		t.Fatalf("expected both clusters, got %d", len(got))
	}
	if len(got[0].Users) != 2 {
		t.Errorf("expected reserved users kept, got %#v", got[0].Users)
	}

	keepers := ClickHouseKeepersToDataSourceModel([]ClickHouseKeeperSpec{{Name: "keeper", Replicas: 3}})
	if len(keepers) != 1 || keepers[0].Replicas.ValueInt64() != 3 {
		t.Errorf("unexpected keepers: %#v", keepers)
	}

	if ClickHouseClustersToDataSourceModel(nil) != nil {
		t.Error("an env without clusters must stay null, not an empty list")
	}
}

func TestClickHouseComputedValuesComeBackFromTheAPI(t *testing.T) {
	// What a create plans: the name is known, every Optional+Computed value is not.
	planned := []ClickHouseClusterModel{{
		Name:         types.StringValue("ch"),
		Image:        types.StringUnknown(),
		InstanceType: types.StringValue("m6i.large"),
		Zones:        types.ListUnknown(types.StringType),
		Shards:       types.Int64Unknown(),
		Replicas:     types.Int64Unknown(),
		Keeper:       types.StringUnknown(),
	}}

	got := ClickHouseClustersToModel(planned, []ClickHouseClusterSpec{minimalClusterSpec("ch")})
	if len(got) != 1 {
		t.Fatalf("expected the planned cluster, got %#v", got)
	}

	c := got[0]
	if c.Image.IsUnknown() || c.Shards.IsUnknown() || c.Replicas.IsUnknown() || c.Keeper.IsUnknown() || c.Zones.IsUnknown() {
		t.Errorf("an unknown left in state fails the post-apply check: %#v", c)
	}
	if c.Image.ValueString() != "altinity/clickhouse-server:25.8" || c.Shards.ValueInt64() != 1 {
		t.Errorf("unexpected resolved values: %#v", c)
	}
	if c.Disk.Iops.ValueInt64() != 3000 || c.Disk.ThroughputMiBs.ValueInt64() != 125 {
		t.Errorf("disk defaults must come back from the API: %#v", c.Disk)
	}
}
