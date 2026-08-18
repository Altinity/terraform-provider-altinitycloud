package hosted_env

import (
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMetricsEndpointToSDK(t *testing.T) {
	t.Run("nil model yields nil input", func(t *testing.T) {
		if got := MetricsEndpointToSDK(nil); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("populates enabled and ranges", func(t *testing.T) {
		got := MetricsEndpointToSDK(&MetricsEndpointModel{
			Enabled:        types.BoolValue(true),
			SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8")},
		})
		if got.Enabled == nil || !*got.Enabled {
			t.Errorf("enabled = %v, want true", got.Enabled)
		}
		if len(got.SourceIPRanges) != 1 || got.SourceIPRanges[0] != "10.0.0.0/8" {
			t.Errorf("source_ip_ranges = %v, want [10.0.0.0/8]", got.SourceIPRanges)
		}
	})

	t.Run("unset ranges stay nil rather than empty", func(t *testing.T) {
		got := MetricsEndpointToSDK(&MetricsEndpointModel{Enabled: types.BoolValue(false)})
		if got.SourceIPRanges != nil {
			t.Errorf("source_ip_ranges = %v, want nil", got.SourceIPRanges)
		}
	})
}

func TestMetricsEndpointToModel(t *testing.T) {
	tests := []struct {
		name     string
		existing *MetricsEndpointModel
		enabled  bool
		ranges   []string
		wantNil  bool
	}{
		// The API always returns the block; state must stay null so an env that
		// never configured it does not drift.
		{name: "unconfigured and disabled stays nil", existing: nil, enabled: false, wantNil: true},
		{name: "unconfigured but enabled out-of-band populates", existing: nil, enabled: true, ranges: []string{"0.0.0.0/0"}},
		{name: "configured stays populated when disabled", existing: &MetricsEndpointModel{}, enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetricsEndpointToModel(tt.existing, tt.enabled, tt.ranges)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("got nil, want a populated model")
			}
			if got.Enabled.ValueBool() != tt.enabled {
				t.Errorf("enabled = %v, want %v", got.Enabled.ValueBool(), tt.enabled)
			}
			if len(got.SourceIPRanges) != len(tt.ranges) {
				t.Errorf("source_ip_ranges = %v, want %v", got.SourceIPRanges, tt.ranges)
			}
		})
	}
}

func TestDatadogToModel(t *testing.T) {
	t.Run("unconfigured and disabled stays nil", func(t *testing.T) {
		if got := DatadogToModel(nil, false, "datadoghq.com", false, false); got != nil {
			t.Errorf("got %v, want nil", got)
		}
	})

	t.Run("unconfigured but enabled out-of-band populates without api key", func(t *testing.T) {
		got := DatadogToModel(nil, true, "us3.datadoghq.com", true, false)
		if got == nil {
			t.Fatal("got nil, want a populated model")
		}
		if !got.EncAPIKey.IsNull() {
			t.Errorf("enc_api_key = %v, want null", got.EncAPIKey)
		}
		if got.Domain.ValueString() != "us3.datadoghq.com" || !got.LogsEnabled.ValueBool() || got.MetricsEnabled.ValueBool() {
			t.Errorf("unexpected model: %+v", got)
		}
	})

	t.Run("configured preserves write-only enc_api_key", func(t *testing.T) {
		existing := &common.DatadogModel{EncAPIKey: types.StringValue("enc-secret")}
		got := DatadogToModel(existing, true, "datadoghq.com", false, true)
		if got.EncAPIKey.ValueString() != "enc-secret" {
			t.Errorf("enc_api_key = %q, want enc-secret", got.EncAPIKey.ValueString())
		}
	})

	t.Run("configured but disabled stays populated", func(t *testing.T) {
		existing := &common.DatadogModel{EncAPIKey: types.StringValue("enc-secret")}
		if got := DatadogToModel(existing, false, "datadoghq.com", false, false); got == nil {
			t.Fatal("got nil, want the block kept in state")
		}
	})
}

func TestCustomDomainsToModel(t *testing.T) {
	t.Run("null prior stays null even when the API echoes an empty list", func(t *testing.T) {
		got, diags := CustomDomainsToModel(types.ListNull(types.StringType), []string{})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !got.IsNull() {
			t.Errorf("got %v, want null", got)
		}
	})

	t.Run("configured prior refreshes from the API", func(t *testing.T) {
		prior := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old.com")})
		got, diags := CustomDomainsToModel(prior, []string{"a.com", "b.com"})
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if len(got.Elements()) != 2 {
			t.Errorf("got %v, want 2 elements", got)
		}
	})
}

func TestNodeGroupKey(t *testing.T) {
	tests := []struct {
		name  string
		group NodeGroupsModel
		want  string
	}{
		{"configured name wins", NodeGroupsModel{Name: types.StringValue("big"), NodeType: types.StringValue("m6i.large")}, "big"},
		{"unknown name falls back to node_type", NodeGroupsModel{Name: types.StringUnknown(), NodeType: types.StringValue("m6i.large")}, "m6i.large"},
		{"null name falls back to node_type", NodeGroupsModel{Name: types.StringNull(), NodeType: types.StringValue("m6i.large")}, "m6i.large"},
		{"empty name falls back to node_type", NodeGroupsModel{Name: types.StringValue(""), NodeType: types.StringValue("m6i.large")}, "m6i.large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NodeGroupKey(tt.group); got != tt.want {
				t.Errorf("NodeGroupKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
