package hosted_env

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
