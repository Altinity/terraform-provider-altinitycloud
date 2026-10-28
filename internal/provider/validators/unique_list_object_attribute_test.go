package validators

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUniqueListObjectAttribute(t *testing.T) {
	t.Parallel()

	attrTypes := map[string]attr.Type{"name": types.StringType}
	elemType := types.ObjectType{AttrTypes: attrTypes}
	object := func(name string) attr.Value {
		return types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue(name)})
	}
	v := UniqueListObjectAttribute("name", "Duplicate", func(name string) string { return "dup " + name })
	run := func(value types.List) diag.Diagnostics {
		resp := &validator.ListResponse{}
		v.ValidateList(context.Background(), validator.ListRequest{Path: path.Root("users"), ConfigValue: value}, resp)
		return resp.Diagnostics
	}

	if d := run(types.ListValueMust(elemType, []attr.Value{object("a"), object("b")})); d.HasError() {
		t.Errorf("unexpected error: %s", d.Errors())
	}
	if d := run(types.ListValueMust(elemType, []attr.Value{object("a"), object("a")})); !d.HasError() || !strings.Contains(d.Errors()[0].Detail(), "dup a") {
		t.Errorf("expected a duplicate error, got %v", d)
	}
	for name, value := range map[string]types.List{"null": types.ListNull(elemType), "unknown": types.ListUnknown(elemType)} {
		if d := run(value); d.HasError() {
			t.Errorf("%s: unexpected error %s", name, d.Errors())
		}
	}
}
