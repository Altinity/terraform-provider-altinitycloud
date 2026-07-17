package env

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReorderByKey(t *testing.T) {
	type item struct {
		Key   string
		Value string
	}

	tests := []struct {
		name          string
		modelKeys     []string
		items         []item
		expectedOrder []string
	}{
		{
			name:      "Preserve model order and add new items",
			modelKeys: []string{"b", "a"},
			items: []item{
				{Key: "a", Value: "1"},
				{Key: "b", Value: "2"},
				{Key: "c", Value: "3"},
			},
			expectedOrder: []string{"b", "a", "c"},
		},
		{
			name:      "All items in model",
			modelKeys: []string{"x", "y"},
			items: []item{
				{Key: "y", Value: "2"},
				{Key: "x", Value: "1"},
			},
			expectedOrder: []string{"x", "y"},
		},
		{
			name:      "Model has keys not in items",
			modelKeys: []string{"a", "missing", "b"},
			items: []item{
				{Key: "b", Value: "2"},
				{Key: "a", Value: "1"},
			},
			expectedOrder: []string{"a", "b"},
		},
		{
			name:      "Empty model",
			modelKeys: []string{},
			items: []item{
				{Key: "a", Value: "1"},
				{Key: "b", Value: "2"},
			},
			expectedOrder: []string{"a", "b"},
		},
		{
			name:          "Empty items",
			modelKeys:     []string{"a"},
			items:         []item{},
			expectedOrder: []string{},
		},
		{
			name:          "Both empty",
			modelKeys:     []string{},
			items:         []item{},
			expectedOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReorderByKey(tt.modelKeys, tt.items,
				func(k string) string { return k },
				func(i item) string { return i.Key },
			)

			if len(result) != len(tt.expectedOrder) {
				t.Fatalf("Expected length %d, got %d", len(tt.expectedOrder), len(result))
			}

			for i, expected := range tt.expectedOrder {
				if result[i].Key != expected {
					t.Errorf("Key at position %d: expected %s, got %s", i, expected, result[i].Key)
				}
			}
		})
	}
}

// []T{} serializes as empty list, not null; nil must stay nil (v0.7.3 regression).
func TestReorderByKeyNilPreserved(t *testing.T) {
	type item struct{ Key string }
	key := func(i item) string { return i.Key }

	t.Run("nil items, empty model", func(t *testing.T) {
		if got := ReorderByKey([]item(nil), []item(nil), key, key); got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("nil items, non-empty model", func(t *testing.T) {
		if got := ReorderByKey([]item{{Key: "a"}}, []item(nil), key, key); got != nil {
			t.Errorf("expected nil, got %#v", got)
		}
	})

	t.Run("non-nil empty items stay non-nil", func(t *testing.T) {
		if got := ReorderByKey([]item{{Key: "a"}}, []item{}, key, key); got == nil {
			t.Error("expected non-nil empty slice, got nil")
		}
	})
}

// Duplicate keys must not collapse to the first match or drop entries.
func TestReorderByKeyDuplicateKeys(t *testing.T) {
	type item struct{ Key, Value string }

	t.Run("empty keys preserved positionally", func(t *testing.T) {
		model := []item{{Key: "", Value: "B"}, {Key: "", Value: "A"}}
		items := []item{{Key: "", Value: "A"}, {Key: "", Value: "B"}}
		result := ReorderByKey(model, items,
			func(m item) string { return m.Key },
			func(s item) string { return s.Key },
		)
		if len(result) != 2 {
			t.Fatalf("expected 2 items, got %d (%v)", len(result), result)
		}
		// Both originals present, none duplicated.
		seen := map[string]int{}
		for _, r := range result {
			seen[r.Value]++
		}
		if seen["A"] != 1 || seen["B"] != 1 {
			t.Fatalf("expected one A and one B, got %v", seen)
		}
	})

	t.Run("shared key matches distinct values", func(t *testing.T) {
		// e.g. two tolerations with key "dedicated" but different values.
		model := []item{{Key: "k", Value: "v2"}, {Key: "k", Value: "v1"}}
		items := []item{{Key: "k", Value: "v1"}, {Key: "k", Value: "v2"}}
		result := ReorderByKey(model, items,
			func(m item) string { return m.Key },
			func(s item) string { return s.Key },
		)
		if len(result) != 2 || result[0].Value == result[1].Value {
			t.Fatalf("expected both distinct values preserved, got %v", result)
		}
	})
}

func TestReorderNodeGroupZones(t *testing.T) {
	type modelNG struct {
		NodeType string
		Zones    types.List
	}
	type apiNG struct {
		NodeType string
		Zones    []string
	}

	zonesList := func(zones ...string) types.List {
		values := make([]attr.Value, len(zones))
		for i, z := range zones {
			values[i] = types.StringValue(z)
		}
		return types.ListValueMust(types.StringType, values)
	}

	run := func(t *testing.T, model []modelNG, items []*apiNG) {
		t.Helper()
		diags := ReorderNodeGroupZones(context.Background(), model, items,
			func(m modelNG) string { return m.NodeType },
			func(s *apiNG) string { return s.NodeType },
			func(m modelNG) types.List { return m.Zones },
			func(s *apiNG) *[]string { return &s.Zones },
		)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
	}

	assertZones := func(t *testing.T, got, want []string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("expected zones %v, got %v", want, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected zones %v, got %v", want, got)
			}
		}
	}

	t.Run("reorders zones to model order per node group", func(t *testing.T) {
		model := []modelNG{
			{NodeType: "system", Zones: zonesList("us-east-1d", "us-east-1a")},
			{NodeType: "user", Zones: zonesList("us-east-1d", "us-east-1a")},
		}
		items := []*apiNG{
			{NodeType: "system", Zones: []string{"us-east-1a", "us-east-1d"}},
			{NodeType: "user", Zones: []string{"us-east-1a", "us-east-1d"}},
		}
		run(t, model, items)
		assertZones(t, items[0].Zones, []string{"us-east-1d", "us-east-1a"})
		assertZones(t, items[1].Zones, []string{"us-east-1d", "us-east-1a"})
	})

	t.Run("null model zones keep API order", func(t *testing.T) {
		model := []modelNG{{NodeType: "system", Zones: types.ListNull(types.StringType)}}
		items := []*apiNG{{NodeType: "system", Zones: []string{"us-east-1b", "us-east-1a"}}}
		run(t, model, items)
		assertZones(t, items[0].Zones, []string{"us-east-1b", "us-east-1a"})
	})

	t.Run("API-only node group untouched", func(t *testing.T) {
		model := []modelNG{{NodeType: "system", Zones: zonesList("us-east-1b", "us-east-1a")}}
		items := []*apiNG{
			{NodeType: "system", Zones: []string{"us-east-1a", "us-east-1b"}},
			{NodeType: "extra", Zones: []string{"us-east-1c", "us-east-1a"}},
		}
		run(t, model, items)
		assertZones(t, items[0].Zones, []string{"us-east-1b", "us-east-1a"})
		assertZones(t, items[1].Zones, []string{"us-east-1c", "us-east-1a"})
	})

	t.Run("duplicate node types pair positionally", func(t *testing.T) {
		model := []modelNG{
			{NodeType: "t4g.large", Zones: zonesList("us-east-1b", "us-east-1a")},
			{NodeType: "t4g.large", Zones: zonesList("us-east-1d", "us-east-1c")},
		}
		items := []*apiNG{
			{NodeType: "t4g.large", Zones: []string{"us-east-1a", "us-east-1b"}},
			{NodeType: "t4g.large", Zones: []string{"us-east-1c", "us-east-1d"}},
		}
		run(t, model, items)
		assertZones(t, items[0].Zones, []string{"us-east-1b", "us-east-1a"})
		assertZones(t, items[1].Zones, []string{"us-east-1d", "us-east-1c"})
	})

	t.Run("extra API zones appended after model order", func(t *testing.T) {
		model := []modelNG{{NodeType: "system", Zones: zonesList("us-east-1d", "us-east-1a")}}
		items := []*apiNG{{NodeType: "system", Zones: []string{"us-east-1a", "us-east-1b", "us-east-1d"}}}
		run(t, model, items)
		assertZones(t, items[0].Zones, []string{"us-east-1d", "us-east-1a", "us-east-1b"})
	})
}

func TestReorderList(t *testing.T) {
	tests := []struct {
		name           string
		model          types.List
		input          []string
		expectedOrder  []string
		expectedLength int
	}{
		{
			name: "Preserve model order and add new items",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("us-east-1a"),
					types.StringValue("us-east-1c"),
				},
			),
			input:          []string{"us-east-1b", "us-east-1a", "us-east-1c", "us-east-1d"},
			expectedOrder:  []string{"us-east-1a", "us-east-1c", "us-east-1b", "us-east-1d"},
			expectedLength: 4,
		},
		{
			name: "All input items exist in model",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("zone1"),
					types.StringValue("zone2"),
					types.StringValue("zone3"),
				},
			),
			input:          []string{"zone3", "zone1", "zone2"},
			expectedOrder:  []string{"zone1", "zone2", "zone3"},
			expectedLength: 3,
		},
		{
			name: "Model has items not in input",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("zone1"),
					types.StringValue("zone2"),
					types.StringValue("missing-zone"),
				},
			),
			input:          []string{"zone2", "zone1", "zone3"},
			expectedOrder:  []string{"zone1", "zone2", "zone3"},
			expectedLength: 3,
		},
		{
			name:           "Empty model with input items",
			model:          types.ListValueMust(types.StringType, []attr.Value{}),
			input:          []string{"zone1", "zone2", "zone3"},
			expectedOrder:  []string{"zone1", "zone2", "zone3"},
			expectedLength: 3,
		},
		{
			name: "Empty input with model items",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("zone1"),
					types.StringValue("zone2"),
				},
			),
			input:          []string{},
			expectedOrder:  []string{},
			expectedLength: 0,
		},
		{
			name:           "Both empty",
			model:          types.ListValueMust(types.StringType, []attr.Value{}),
			input:          []string{},
			expectedOrder:  []string{},
			expectedLength: 0,
		},
		{
			name: "Model order preservation with multiple new items",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("eu-west-1a"),
					types.StringValue("eu-west-1c"),
				},
			),
			input:          []string{"eu-west-1d", "eu-west-1b", "eu-west-1a", "eu-west-1e", "eu-west-1c"},
			expectedOrder:  []string{"eu-west-1a", "eu-west-1c", "eu-west-1d", "eu-west-1b", "eu-west-1e"},
			expectedLength: 5,
		},
		{
			name: "Single item in model and input",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("single-zone"),
				},
			),
			input:          []string{"single-zone"},
			expectedOrder:  []string{"single-zone"},
			expectedLength: 1,
		},
		{
			name: "Complex scenario with duplicates in input",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("priority-zone"),
					types.StringValue("secondary-zone"),
				},
			),
			input:          []string{"new-zone", "priority-zone", "secondary-zone", "priority-zone", "another-zone"},
			expectedOrder:  []string{"priority-zone", "secondary-zone", "new-zone", "another-zone"},
			expectedLength: 4,
		},
		{
			name: "Model with reverse order compared to input",
			model: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("zone-c"),
					types.StringValue("zone-b"),
					types.StringValue("zone-a"),
				},
			),
			input:          []string{"zone-a", "zone-b", "zone-c"},
			expectedOrder:  []string{"zone-c", "zone-b", "zone-a"},
			expectedLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := ReorderList(context.Background(), tt.model, tt.input)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if len(result) != tt.expectedLength {
				t.Errorf("Expected length %d, got %d", tt.expectedLength, len(result))
			}

			if len(result) != len(tt.expectedOrder) {
				t.Fatalf("Result length %d doesn't match expected order length %d", len(result), len(tt.expectedOrder))
			}

			for i, expected := range tt.expectedOrder {
				if result[i] != expected {
					t.Errorf("Item at position %d: expected '%s', got '%s'", i, expected, result[i])
				}
			}

			seen := make(map[string]bool)
			for _, item := range result {
				if seen[item] {
					t.Errorf("Duplicate item found in result: '%s'", item)
				}
				seen[item] = true
			}

			inputMap := make(map[string]bool)
			for _, item := range tt.input {
				inputMap[item] = true
			}

			for _, resultItem := range result {
				if !inputMap[resultItem] {
					t.Errorf("Result contains item '%s' that was not in input", resultItem)
				}
			}
		})
	}
}
