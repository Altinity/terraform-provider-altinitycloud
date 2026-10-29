package env

import (
	"math"
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
		prior := map[string]ClickHouseClusterModel{"mine": {}}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{
			minimalClusterSpec("mine"),
			minimalClusterSpec("made-in-the-console"),
		})

		if _, ok := got["mine"]; len(got) != 1 || !ok {
			t.Errorf("expected only the managed cluster, got %#v", got)
		}
	})

	t.Run("a managed cluster deleted elsewhere drops out of state", func(t *testing.T) {
		prior := map[string]ClickHouseClusterModel{"mine": {}, "gone": {}}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{minimalClusterSpec("mine")})

		if _, ok := got["gone"]; len(got) != 1 || ok {
			t.Errorf("expected only mine, got %#v", got)
		}
	})

	t.Run("empty API lists stay null, never empty maps", func(t *testing.T) {
		prior := map[string]ClickHouseClusterModel{"ch": {}}
		got := ClickHouseClustersToModel(prior, []ClickHouseClusterSpec{minimalClusterSpec("ch")})

		c := got["ch"]
		if c.ExtraDisks != nil || c.Users != nil || c.Profiles != nil || c.Quotas != nil || c.ServerSettings != nil {
			t.Errorf("omitted collections must stay nil, got %#v", c)
		}
	})
}

func TestClickHouseUsersToModel(t *testing.T) {
	t.Run("reserved users are dropped", func(t *testing.T) {
		prior := map[string]ClickHouseUserModel{"app": {}}
		got := usersToModel(prior, []ClickHouseUserSpec{
			{Name: "app", PasswordType: "SHA256_HEX"},
			{Name: "grafana", PasswordType: "SHA256_HEX"},
			{Name: "datadog", PasswordType: "SHA256_HEX"},
		})

		if _, ok := got["app"]; len(got) != 1 || !ok {
			t.Errorf("expected only app, got %#v", got)
		}
	})

	t.Run("enc_value comes from prior state, never from the API", func(t *testing.T) {
		prior := map[string]ClickHouseUserModel{"app": {
			Password: &ClickHouseUserPasswordModel{
				Type:     types.StringValue("SHA256_HEX"),
				EncValue: types.StringValue("kept-from-config"),
			},
		}}
		got := usersToModel(prior, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})

		if got["app"].Password.EncValue.ValueString() != "kept-from-config" {
			t.Errorf("enc_value = %v, want the configured value", got["app"].Password.EncValue)
		}
	})

	t.Run("a user absent from prior state has a null enc_value", func(t *testing.T) {
		got := usersToModel(map[string]ClickHouseUserModel{}, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})

		if !got["app"].Password.EncValue.IsNull() {
			t.Errorf("enc_value = %v, want null", got["app"].Password.EncValue)
		}
	})

	t.Run("an omitted networks block stays nil", func(t *testing.T) {
		prior := map[string]ClickHouseUserModel{"app": {}}
		got := usersToModel(prior, []ClickHouseUserSpec{{Name: "app", PasswordType: "SHA256_HEX"}})["app"]

		if got.Networks != nil {
			t.Errorf("networks = %#v, want nil", got.Networks)
		}
		if !got.Databases.IsNull() || !got.Grants.IsNull() {
			t.Errorf("omitted lists must stay null, got %#v / %#v", got.Databases, got.Grants)
		}
	})

	t.Run("ips follow config order", func(t *testing.T) {
		prior := map[string]ClickHouseUserModel{"app": {
			Networks: &ClickHouseUserNetworksModel{Ips: testList(t, "10.0.0.0/8", "192.168.0.0/16")},
		}}
		got := usersToModel(prior, []ClickHouseUserSpec{{
			Name:         "app",
			Ips:          []string{"192.168.0.0/16", "10.0.0.0/8"},
			PasswordType: "SHA256_HEX",
		}})

		ips := listStrings(t, got["app"].Networks.Ips)
		if len(ips) != 2 || ips[0] != "10.0.0.0/8" {
			t.Errorf("expected config order, got %v", ips)
		}
	})
}

func TestClickHouseQuotasToModel(t *testing.T) {
	queries := uint64(1000)
	prior := map[string]ClickHouseQuotaModel{"q": {}}

	got := quotasToModel(prior, []ClickHouseQuotaSpec{{
		Name:    "q",
		KeyedBy: "IP_ADDRESS",
		Intervals: []ClickHouseQuotaIntervalSpec{{
			DurationSeconds: 3600,
			Queries:         &queries,
		}},
	}})

	q := got["q"]
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
		prior := map[string]ClickHouseKeeperModel{"keeper": {
			Zones: testList(t, "us-east-1b", "us-east-1a"),
		}}
		got := ClickHouseKeepersToModel(prior, []ClickHouseKeeperSpec{{
			Name:     "keeper",
			Zones:    []string{"us-east-1a", "us-east-1b"},
			Replicas: 3,
			Disk:     ClickHouseDiskSpec{SizeGiB: 30},
		}})

		zones := listStrings(t, got["keeper"].Zones)
		if zones[0] != "us-east-1b" {
			t.Errorf("expected config order, got %v", zones)
		}
		if got["keeper"].Replicas.ValueInt64() != 3 {
			t.Errorf("replicas = %v, want 3", got["keeper"].Replicas)
		}
	})
}

func TestClickHouseDataSourceModelReportsEverything(t *testing.T) {
	specs := []ClickHouseClusterSpec{minimalClusterSpec("mine"), minimalClusterSpec("made-in-the-console")}
	specs[0].Users = []ClickHouseUserSpec{
		{Name: "app", PasswordType: "SHA256_HEX"},
		{Name: "grafana", PasswordType: "SHA256_HEX"},
	}

	// Without prior state the resource owns nothing; the data source still reports everything.
	if got := ClickHouseClustersToModel(nil, specs); got != nil {
		t.Errorf("resource mapper adopted %#v", got)
	}

	got := ClickHouseClustersToDataSourceModel(specs)
	if len(got) != 2 {
		t.Fatalf("expected both clusters, got %d", len(got))
	}
	if len(got["mine"].Users) != 2 {
		t.Errorf("expected reserved users kept, got %#v", got["mine"].Users)
	}

	keepers := ClickHouseKeepersToDataSourceModel([]ClickHouseKeeperSpec{{Name: "keeper", Replicas: 3}})
	if len(keepers) != 1 || keepers["keeper"].Replicas.ValueInt64() != 3 {
		t.Errorf("unexpected keepers: %#v", keepers)
	}

	if ClickHouseClustersToDataSourceModel(nil) != nil {
		t.Error("an env without clusters must stay null, not an empty map")
	}
}

func TestClickHouseComputedValuesComeBackFromTheAPI(t *testing.T) {
	// What a create plans: the key is known, every Optional+Computed value is not.
	planned := map[string]ClickHouseClusterModel{"ch": {
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

	c := got["ch"]
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

func TestClickHouseEmptyAPIEchoesMirrorThePlan(t *testing.T) {
	// JSON [] unmarshals to a non-nil empty slice; it must read back as whatever the plan held.
	spec := minimalClusterSpec("ch")
	spec.ExtraDisks = []ClickHouseNamedDiskSpec{}
	spec.ServerSettings = []ClickHouseSettingSpec{}
	spec.Users = []ClickHouseUserSpec{
		{Name: "grafana", PasswordType: "SHA256_HEX"},
		{Name: "datadog", PasswordType: "SHA256_HEX"},
	}

	omitted := ClickHouseClustersToModel(map[string]ClickHouseClusterModel{"ch": {}}, []ClickHouseClusterSpec{spec})["ch"]
	if omitted.ExtraDisks != nil || omitted.ServerSettings != nil || omitted.Users != nil {
		t.Errorf("omitted collections must stay nil, got %#v", omitted)
	}

	declaredEmpty := ClickHouseClustersToModel(map[string]ClickHouseClusterModel{"ch": {
		ExtraDisks:     map[string]ClickHouseExtraDiskModel{},
		ServerSettings: []ClickHouseSettingModel{},
		Users:          map[string]ClickHouseUserModel{},
	}}, []ClickHouseClusterSpec{spec})["ch"]
	if declaredEmpty.ExtraDisks == nil || declaredEmpty.ServerSettings == nil || declaredEmpty.Users == nil {
		t.Errorf("empty collections in the plan must read back empty, not null, got %#v", declaredEmpty)
	}
	if len(declaredEmpty.ExtraDisks)+len(declaredEmpty.ServerSettings)+len(declaredEmpty.Users) != 0 {
		t.Errorf("expected empty collections, got %#v", declaredEmpty)
	}

	priorUser := map[string]ClickHouseUserModel{"app": {}}
	u := usersToModel(priorUser, []ClickHouseUserSpec{{Name: "app", Ips: []string{}, PasswordType: "SHA256_HEX"}})["app"]
	if u.Networks != nil {
		t.Errorf("an empty ips echo must not invent a networks block, got %#v", u.Networks)
	}
	if !u.Grants.IsNull() {
		t.Errorf("grants was null in the plan, so [] must come back null, got %#v", u.Grants)
	}

	withEmpty := map[string]ClickHouseUserModel{"app": {
		Networks: &ClickHouseUserNetworksModel{Ips: types.ListNull(types.StringType)},
		Grants:   testList(t),
	}}
	u = usersToModel(withEmpty, []ClickHouseUserSpec{{Name: "app", Ips: []string{}, PasswordType: "SHA256_HEX"}})["app"]
	if u.Networks == nil || !u.Networks.Ips.IsNull() {
		t.Errorf("ips was null in the plan, so [] must come back null, got %#v", u.Networks)
	}
	if u.Grants.IsNull() || len(u.Grants.Elements()) != 0 {
		t.Errorf("grants was [] in the plan, so [] must come back empty, got %#v", u.Grants)
	}
}

func TestAdoptClickHouseClusters(t *testing.T) {
	specs := []ClickHouseClusterSpec{minimalClusterSpec("mine"), minimalClusterSpec("made-in-the-console")}
	specs[1].Users = []ClickHouseUserSpec{
		{Name: "app", PasswordType: "SHA256_HEX"},
		{Name: "grafana", PasswordType: "SHA256_HEX"},
	}

	got := AdoptClickHouseClusters(specs)
	console, ok := got["made-in-the-console"]
	if len(got) != 2 || !ok {
		t.Fatalf("expected both clusters adopted, got %#v", got)
	}
	app, ok := console.Users["app"]
	if len(console.Users) != 1 || !ok {
		t.Errorf("reserved users must not be adopted, got %#v", console.Users)
	}
	if !app.Password.EncValue.IsNull() {
		t.Errorf("an adopted user has no enc_value to keep, got %v", app.Password.EncValue)
	}
	if AdoptClickHouseClusters(nil) != nil {
		t.Error("an env without clusters must stay null")
	}

	keepers := AdoptClickHouseKeepers([]ClickHouseKeeperSpec{{Name: "keeper", Replicas: 3}})
	if len(keepers) != 1 || keepers["keeper"].Replicas.ValueInt64() != 3 {
		t.Errorf("unexpected keepers: %#v", keepers)
	}
}

func TestClickHouseQuotaIntervalsFollowConfigOrder(t *testing.T) {
	prior := map[string]ClickHouseQuotaModel{"q": {
		Intervals: []ClickHouseQuotaIntervalModel{
			{DurationSeconds: types.Int64Value(86400)},
			{DurationSeconds: types.Int64Value(3600)},
		},
	}}

	got := quotasToModel(prior, []ClickHouseQuotaSpec{{
		Name: "q",
		Intervals: []ClickHouseQuotaIntervalSpec{
			{DurationSeconds: 3600},
			{DurationSeconds: 86400},
		},
	}})

	intervals := got["q"].Intervals
	if intervals[0].DurationSeconds.ValueInt64() != 86400 || intervals[1].DurationSeconds.ValueInt64() != 3600 {
		t.Errorf("expected config order [86400 3600], got %#v", intervals)
	}
}

func TestKeeperToModel(t *testing.T) {
	if got := keeperToModel(""); !got.IsNull() {
		t.Errorf("the API's empty keeper must read back as null, got %v", got)
	}
	if got := keeperToModel("keeper"); got.ValueString() != "keeper" {
		t.Errorf("expected keeper, got %v", got)
	}
}

func TestValidateClickHouseQuotaLimits(t *testing.T) {
	fits, overflows := uint64(math.MaxInt64), uint64(math.MaxInt64)+1
	spec := func(v uint64) []ClickHouseClusterSpec {
		return []ClickHouseClusterSpec{{
			Name:   "ch",
			Quotas: []ClickHouseQuotaSpec{{Name: "q", Intervals: []ClickHouseQuotaIntervalSpec{{DurationSeconds: 3600, ReadRows: &v}}}},
		}}
	}

	if diags := ValidateClickHouseQuotaLimits(spec(fits)); diags.HasError() {
		t.Errorf("MaxInt64 fits in Int64, got %v", diags)
	}
	// Set in the console or through the API, a UInt64 above MaxInt64 would read back negative.
	if diags := ValidateClickHouseQuotaLimits(spec(overflows)); len(diags.Errors()) != 1 {
		t.Errorf("expected one out-of-range error, got %v", diags)
	}
}
