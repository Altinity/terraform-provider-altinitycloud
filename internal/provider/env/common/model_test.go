package env

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
			result := ReorderList(tt.model, tt.input)

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
