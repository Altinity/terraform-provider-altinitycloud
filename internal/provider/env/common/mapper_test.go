package env

import (
	"context"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCustomDomainsToSDK(t *testing.T) {
	ctx := context.Background()
	list := func(vals ...string) types.List {
		elems := make([]string, len(vals))
		copy(elems, vals)
		l, diags := ListToModel(elems)
		if diags.HasError() {
			t.Fatalf("ListToModel: %v", diags)
		}
		return l
	}

	tests := []struct {
		name          string
		customDomain  types.String
		customDomains types.List
		wantDomain    *string
		wantDomains   []string
		wantErr       bool
	}{
		{
			name:          "list set, scalar null -> list wins, scalar nil",
			customDomain:  types.StringNull(),
			customDomains: list("a.com", "b.com"),
			wantDomain:    nil,
			wantDomains:   []string{"a.com", "b.com"},
		},
		{
			name:          "scalar set, list null -> scalar wins",
			customDomain:  types.StringValue("a.com"),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    strPtr("a.com"),
			wantDomains:   nil,
		},
		{
			name:          "neither set (both null) -> both nil",
			customDomain:  types.StringNull(),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    nil,
			wantDomains:   nil,
		},
		{
			name:          "scalar unknown, list null -> treated as not set",
			customDomain:  types.StringUnknown(),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    nil,
			wantDomains:   nil,
		},
		{
			name:          "list unknown, scalar set -> scalar wins",
			customDomain:  types.StringValue("a.com"),
			customDomains: types.ListUnknown(types.StringType),
			wantDomain:    strPtr("a.com"),
			wantDomains:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, gotDomains, diags := CustomDomainsToSDK(ctx, tt.customDomain, tt.customDomains)
			if diags.HasError() != tt.wantErr {
				t.Fatalf("diags error = %v, want %v (%v)", diags.HasError(), tt.wantErr, diags)
			}
			if (gotDomain == nil) != (tt.wantDomain == nil) {
				t.Fatalf("domain nil mismatch: got %v, want %v", gotDomain, tt.wantDomain)
			}
			if gotDomain != nil && *gotDomain != *tt.wantDomain {
				t.Fatalf("domain = %q, want %q", *gotDomain, *tt.wantDomain)
			}
			if len(gotDomains) != len(tt.wantDomains) {
				t.Fatalf("domains = %v, want %v", gotDomains, tt.wantDomains)
			}
			for i := range gotDomains {
				if gotDomains[i] != tt.wantDomains[i] {
					t.Fatalf("domains[%d] = %q, want %q", i, gotDomains[i], tt.wantDomains[i])
				}
			}
		})
	}
}

func TestCustomDomainsToModel(t *testing.T) {
	list := func(vals ...string) types.List {
		l, diags := ListToModel(vals)
		if diags.HasError() {
			t.Fatalf("ListToModel: %v", diags)
		}
		return l
	}

	tests := []struct {
		name           string
		prior          types.List
		specDomain     *string
		specDomains    []string
		wantDomainNull bool
		wantDomain     string
		wantListNull   bool
		wantList       []string
	}{
		{
			name:           "list-managed (prior list set) -> refresh list, scalar null",
			prior:          list("a.com", "b.com"),
			specDomain:     strPtr("a.com"),
			specDomains:    []string{"a.com", "b.com"},
			wantDomainNull: true,
			wantListNull:   false,
			wantList:       []string{"a.com", "b.com"},
		},
		{
			name:           "scalar-managed (prior list null) -> mirror scalar, list null",
			prior:          types.ListNull(types.StringType),
			specDomain:     strPtr("a.com"),
			specDomains:    []string{"a.com"},
			wantDomainNull: false,
			wantDomain:     "a.com",
			wantListNull:   true,
		},
		{
			name:           "neither set -> both null",
			prior:          types.ListNull(types.StringType),
			specDomain:     nil,
			specDomains:    nil,
			wantDomainNull: true,
			wantListNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, gotList, diags := CustomDomainsToModel(tt.prior, tt.specDomain, tt.specDomains)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if gotDomain.IsNull() != tt.wantDomainNull {
				t.Fatalf("domain null = %v, want %v (val %q)", gotDomain.IsNull(), tt.wantDomainNull, gotDomain.ValueString())
			}
			if !tt.wantDomainNull && gotDomain.ValueString() != tt.wantDomain {
				t.Fatalf("domain = %q, want %q", gotDomain.ValueString(), tt.wantDomain)
			}
			if gotList.IsNull() != tt.wantListNull {
				t.Fatalf("list null = %v, want %v", gotList.IsNull(), tt.wantListNull)
			}
			if !tt.wantListNull {
				var got []string
				gotList.ElementsAs(context.Background(), &got, false)
				if len(got) != len(tt.wantList) {
					t.Fatalf("list = %v, want %v", got, tt.wantList)
				}
				for i := range got {
					if got[i] != tt.wantList[i] {
						t.Fatalf("list[%d] = %q, want %q", i, got[i], tt.wantList[i])
					}
				}
			}
		})
	}
}

func TestDatadogToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    *DatadogModel
		expected *client.DatadogSpecInput
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Full config",
			input: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("us3.datadoghq.com"),
				LogsEnabled:    types.BoolValue(true),
				MetricsEnabled: types.BoolValue(true),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(true),
				EncAPIKey:      strPtr("enc-secret"),
				Domain:         strPtr("us3.datadoghq.com"),
				LogsEnabled:    boolPtr(true),
				MetricsEnabled: boolPtr(true),
			},
		},
		{
			name: "Enc API key unset is omitted",
			input: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(true),
				EncAPIKey:      nil,
				Domain:         strPtr("datadoghq.com"),
				LogsEnabled:    boolPtr(false),
				MetricsEnabled: boolPtr(false),
			},
		},
		{
			name: "Disabled",
			input: &DatadogModel{
				Enabled:        types.BoolValue(false),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
			expected: &client.DatadogSpecInput{
				Enabled:        boolPtr(false),
				EncAPIKey:      nil,
				Domain:         strPtr("datadoghq.com"),
				LogsEnabled:    boolPtr(false),
				MetricsEnabled: boolPtr(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DatadogToSDK(tt.input)

			if (tt.expected == nil) != (result == nil) {
				t.Fatalf("Expected nil: %v, got nil: %v", tt.expected == nil, result == nil)
			}
			if tt.expected == nil {
				return
			}

			assertBoolPtr(t, "Enabled", tt.expected.Enabled, result.Enabled)
			assertStrPtr(t, "EncAPIKey", tt.expected.EncAPIKey, result.EncAPIKey)
			assertStrPtr(t, "Domain", tt.expected.Domain, result.Domain)
			assertBoolPtr(t, "LogsEnabled", tt.expected.LogsEnabled, result.LogsEnabled)
			assertBoolPtr(t, "MetricsEnabled", tt.expected.MetricsEnabled, result.MetricsEnabled)
		})
	}
}

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

func assertBoolPtr(t *testing.T, field string, expected, got *bool) {
	t.Helper()
	if (expected == nil) != (got == nil) {
		t.Errorf("%s nil mismatch: expected nil %v, got nil %v", field, expected == nil, got == nil)
		return
	}
	if expected != nil && *expected != *got {
		t.Errorf("%s mismatch: expected %v, got %v", field, *expected, *got)
	}
}

func assertStrPtr(t *testing.T, field string, expected, got *string) {
	t.Helper()
	if (expected == nil) != (got == nil) {
		t.Errorf("%s nil mismatch: expected nil %v, got nil %v", field, expected == nil, got == nil)
		return
	}
	if expected != nil && *expected != *got {
		t.Errorf("%s mismatch: expected %q, got %q", field, *expected, *got)
	}
}

func TestMetricsEndpointToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricsEndpointModel
		expected *client.MetricsEndpointSpecInput
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "enabled with source ip ranges",
			input: &MetricsEndpointModel{
				Enabled:        types.BoolValue(true),
				SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8"), types.StringValue("192.168.1.0/24")},
			},
			expected: &client.MetricsEndpointSpecInput{
				Enabled:        boolPtr(true),
				SourceIPRanges: []string{"10.0.0.0/8", "192.168.1.0/24"},
			},
		},
		{
			name: "disabled with empty source ip ranges",
			input: &MetricsEndpointModel{
				Enabled:        types.BoolValue(false),
				SourceIPRanges: []types.String{},
			},
			expected: &client.MetricsEndpointSpecInput{
				Enabled:        boolPtr(false),
				SourceIPRanges: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetricsEndpointToSDK(tt.input)

			if (tt.expected == nil) != (got == nil) {
				t.Fatalf("expected nil: %v, got nil: %v", tt.expected == nil, got == nil)
			}
			if tt.expected == nil {
				return
			}

			assertBoolPtr(t, "Enabled", tt.expected.Enabled, got.Enabled)
			assertStrings(t, "SourceIPRanges", tt.expected.SourceIPRanges, got.SourceIPRanges)
		})
	}
}

func TestMetricsEndpointToModel(t *testing.T) {
	tests := []struct {
		name           string
		existing       *MetricsEndpointModel
		enabled        bool
		sourceIPRanges []string
		expected       *MetricsEndpointModel
	}{
		{
			name:           "enabled populates the block",
			enabled:        true,
			sourceIPRanges: []string{"10.0.0.0/8", "172.16.0.0/12"},
			expected: &MetricsEndpointModel{
				Enabled:        types.BoolValue(true),
				SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8"), types.StringValue("172.16.0.0/12")},
			},
		},
		{
			name:           "disabled and unconfigured stays null",
			enabled:        false,
			sourceIPRanges: []string{},
			expected:       nil,
		},
		{
			name:           "disabled but previously configured is preserved",
			existing:       &MetricsEndpointModel{Enabled: types.BoolValue(false)},
			enabled:        false,
			sourceIPRanges: []string{},
			expected: &MetricsEndpointModel{
				Enabled:        types.BoolValue(false),
				SourceIPRanges: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetricsEndpointToModel(tt.existing, tt.enabled, tt.sourceIPRanges)

			if (tt.expected == nil) != (got == nil) {
				t.Fatalf("expected nil: %v, got nil: %v", tt.expected == nil, got == nil)
			}
			if tt.expected == nil {
				return
			}

			if tt.expected.Enabled.ValueBool() != got.Enabled.ValueBool() {
				t.Errorf("Enabled: expected %v, got %v", tt.expected.Enabled.ValueBool(), got.Enabled.ValueBool())
			}
			if len(tt.expected.SourceIPRanges) != len(got.SourceIPRanges) {
				t.Fatalf("SourceIPRanges count: expected %d, got %d", len(tt.expected.SourceIPRanges), len(got.SourceIPRanges))
			}
			for i, expected := range tt.expected.SourceIPRanges {
				if expected.ValueString() != got.SourceIPRanges[i].ValueString() {
					t.Errorf("SourceIPRanges[%d]: expected %q, got %q", i, expected.ValueString(), got.SourceIPRanges[i].ValueString())
				}
			}
		})
	}
}

func TestDatadogToModel(t *testing.T) {
	tests := []struct {
		name           string
		existing       *DatadogModel
		enabled        bool
		domain         string
		logsEnabled    bool
		metricsEnabled bool
		expected       *DatadogModel
	}{
		{
			// Existing envs upgrading without a datadog block must not drift: the API
			// always returns the block (disabled by default) but state stays null.
			name:     "unconfigured and disabled stays nil",
			domain:   "datadoghq.com",
			expected: nil,
		},
		{
			name:        "unconfigured but enabled out-of-band populates without api key",
			enabled:     true,
			domain:      "us3.datadoghq.com",
			logsEnabled: true,
			expected: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringNull(),
				Domain:         types.StringValue("us3.datadoghq.com"),
				LogsEnabled:    types.BoolValue(true),
				MetricsEnabled: types.BoolValue(false),
			},
		},
		{
			name:           "configured preserves write-only enc_api_key",
			existing:       &DatadogModel{EncAPIKey: types.StringValue("enc-secret")},
			enabled:        true,
			domain:         "datadoghq.com",
			metricsEnabled: true,
			expected: &DatadogModel{
				Enabled:        types.BoolValue(true),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(true),
			},
		},
		{
			name: "configured but disabled stays populated",
			existing: &DatadogModel{
				Enabled:   types.BoolValue(true),
				EncAPIKey: types.StringValue("enc-secret"),
			},
			domain: "datadoghq.com",
			expected: &DatadogModel{
				Enabled:        types.BoolValue(false),
				EncAPIKey:      types.StringValue("enc-secret"),
				Domain:         types.StringValue("datadoghq.com"),
				LogsEnabled:    types.BoolValue(false),
				MetricsEnabled: types.BoolValue(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DatadogToModel(tt.existing, tt.enabled, tt.domain, tt.logsEnabled, tt.metricsEnabled)

			if (tt.expected == nil) != (got == nil) {
				t.Fatalf("expected nil: %v, got nil: %v", tt.expected == nil, got == nil)
			}
			if tt.expected == nil {
				return
			}

			if tt.expected.Enabled.ValueBool() != got.Enabled.ValueBool() {
				t.Errorf("Enabled: expected %v, got %v", tt.expected.Enabled.ValueBool(), got.Enabled.ValueBool())
			}
			if tt.expected.EncAPIKey.IsNull() != got.EncAPIKey.IsNull() {
				t.Errorf("EncAPIKey null: expected %v, got %v", tt.expected.EncAPIKey.IsNull(), got.EncAPIKey.IsNull())
			}
			if tt.expected.EncAPIKey.ValueString() != got.EncAPIKey.ValueString() {
				t.Errorf("EncAPIKey: expected %q, got %q", tt.expected.EncAPIKey.ValueString(), got.EncAPIKey.ValueString())
			}
			if tt.expected.Domain.ValueString() != got.Domain.ValueString() {
				t.Errorf("Domain: expected %q, got %q", tt.expected.Domain.ValueString(), got.Domain.ValueString())
			}
			if tt.expected.LogsEnabled.ValueBool() != got.LogsEnabled.ValueBool() {
				t.Errorf("LogsEnabled: expected %v, got %v", tt.expected.LogsEnabled.ValueBool(), got.LogsEnabled.ValueBool())
			}
			if tt.expected.MetricsEnabled.ValueBool() != got.MetricsEnabled.ValueBool() {
				t.Errorf("MetricsEnabled: expected %v, got %v", tt.expected.MetricsEnabled.ValueBool(), got.MetricsEnabled.ValueBool())
			}
		})
	}
}

func TestMaintenanceWindowsToModel(t *testing.T) {
	type window struct {
		name          string
		enabled       bool
		hour          int64
		lengthInHours int64
		days          []string
	}

	tests := []struct {
		name     string
		input    []*client.AWSEnvSpecFragment_MaintenanceWindows
		expected []window
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []*client.AWSEnvSpecFragment_MaintenanceWindows{},
			expected: nil,
		},
		{
			name: "single window",
			input: []*client.AWSEnvSpecFragment_MaintenanceWindows{
				{Name: "nightly-backup", Enabled: true, Hour: 3, LengthInHours: 2, Days: []client.Day{"friday"}},
			},
			expected: []window{
				{name: "nightly-backup", enabled: true, hour: 3, lengthInHours: 2, days: []string{"friday"}},
			},
		},
		{
			name: "multiple windows keep order",
			input: []*client.AWSEnvSpecFragment_MaintenanceWindows{
				{Name: "weekly-maintenance", Hour: 2, LengthInHours: 4, Days: []client.Day{"saturday", "sunday"}},
				{Name: "daily-maintenance", Hour: 1, LengthInHours: 1, Days: []client.Day{"monday", "tuesday", "wednesday"}},
			},
			expected: []window{
				{name: "weekly-maintenance", hour: 2, lengthInHours: 4, days: []string{"saturday", "sunday"}},
				{name: "daily-maintenance", hour: 1, lengthInHours: 1, days: []string{"monday", "tuesday", "wednesday"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaintenanceWindowsToModel(tt.input)

			// nil rather than empty: an empty list would show up as a diff.
			if tt.expected == nil {
				if got != nil {
					t.Fatalf("expected nil, got %d windows", len(got))
				}
				return
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d windows, got %d", len(tt.expected), len(got))
			}

			for i, expected := range tt.expected {
				if got[i].Name.ValueString() != expected.name {
					t.Errorf("[%d] Name: expected %q, got %q", i, expected.name, got[i].Name.ValueString())
				}
				if got[i].Enabled.ValueBool() != expected.enabled {
					t.Errorf("[%d] Enabled: expected %v, got %v", i, expected.enabled, got[i].Enabled.ValueBool())
				}
				if got[i].Hour.ValueInt64() != expected.hour {
					t.Errorf("[%d] Hour: expected %d, got %d", i, expected.hour, got[i].Hour.ValueInt64())
				}
				if got[i].LengthInHours.ValueInt64() != expected.lengthInHours {
					t.Errorf("[%d] LengthInHours: expected %d, got %d", i, expected.lengthInHours, got[i].LengthInHours.ValueInt64())
				}
				assertDays(t, i, expected.days, got[i].Days)
			}
		})
	}
}

// A nil element must not panic: the generated getters are nil-safe, so it maps to
// a zero-valued window.
func TestMaintenanceWindowsToModelNilElement(t *testing.T) {
	got := MaintenanceWindowsToModel([]*client.AWSEnvSpecFragment_MaintenanceWindows{nil})

	if len(got) != 1 {
		t.Fatalf("expected 1 window, got %d", len(got))
	}
	if got[0].Name.ValueString() != "" {
		t.Errorf("expected an empty name, got %q", got[0].Name.ValueString())
	}
}

func assertStrings(t *testing.T, field string, expected, got []string) {
	t.Helper()

	if len(expected) != len(got) {
		t.Fatalf("%s count: expected %d, got %d", field, len(expected), len(got))
	}
	for i := range expected {
		if expected[i] != got[i] {
			t.Errorf("%s[%d]: expected %q, got %q", field, i, expected[i], got[i])
		}
	}
}

func assertDays(t *testing.T, index int, expected []string, got []types.String) {
	t.Helper()

	if len(expected) != len(got) {
		t.Fatalf("[%d] Days count: expected %d, got %d", index, len(expected), len(got))
	}
	for i := range expected {
		if expected[i] != got[i].ValueString() {
			t.Errorf("[%d] Days[%d]: expected %q, got %q", index, i, expected[i], got[i].ValueString())
		}
	}
}
