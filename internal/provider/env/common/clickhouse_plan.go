package env

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Terraform pairs list entries by index, so a rename or insert hands a neighbour's computed values to the new entry.
var clickHouseListKeys = map[string]string{
	"clickhouse_clusters": "name",
	"clickhouse_keepers":  "name",
	"extra_disks":         "name",
	"users":               "name",
	"profiles":            "name",
	"quotas":              "name",
	"settings":            "key",
	"server_settings":     "key",
	"intervals":           "duration",
}

// Optional+Computed values the config omits plan as the same-keyed state entry's value, or unknown when that entry is new.
func UseStateForUnknownByKey(ctx context.Context, name string, schema rschema.Attribute, plan, state, config attr.Value) attr.Value {
	switch s := schema.(type) {
	case rschema.ListNestedAttribute:
		p, ok := plan.(types.List)
		if !ok || p.IsNull() || p.IsUnknown() {
			return plan
		}

		key := clickHouseListKeys[name]
		stateByKey := make(map[string]attr.Value)
		if l, ok := state.(types.List); ok && !l.IsNull() && !l.IsUnknown() {
			for _, e := range l.Elements() {
				if k, ok := elementKey(e, key); ok {
					stateByKey[k] = e
				}
			}
		}

		elements := make([]attr.Value, 0, len(p.Elements()))
		for i, e := range p.Elements() {
			var prior attr.Value
			if k, ok := elementKey(e, key); ok {
				prior = stateByKey[k]
			}
			elements = append(elements, useStateForObject(ctx, s.NestedObject.Attributes, e, prior, configElement(config, i)))
		}
		return types.ListValueMust(p.ElementType(ctx), elements)
	case rschema.SingleNestedAttribute:
		return useStateForObject(ctx, s.Attributes, plan, state, config)
	}

	if !isNull(config) || !computedWithoutDefault(schema) {
		return plan
	}
	if state == nil {
		return unknownOf(ctx, plan)
	}
	return state
}

func UseStateForUnknownByKeyList(ctx context.Context, name string, schema rschema.ListNestedAttribute, plan, state, config types.List) types.List {
	if filled, ok := UseStateForUnknownByKey(ctx, name, schema, plan, state, config).(types.List); ok {
		return filled
	}
	return plan
}

func useStateForObject(ctx context.Context, attributes map[string]rschema.Attribute, plan, state, config attr.Value) attr.Value {
	p, ok := plan.(types.Object)
	if !ok || p.IsNull() || p.IsUnknown() {
		return plan
	}

	attrs := make(map[string]attr.Value, len(p.Attributes()))
	for name, value := range p.Attributes() {
		attrs[name] = UseStateForUnknownByKey(ctx, name, attributes[name], value, stateAttribute(state, name), configAttribute(config, name))
	}
	return types.ObjectValueMust(p.AttributeTypes(ctx), attrs)
}

func computedWithoutDefault(a rschema.Attribute) bool {
	if a == nil || !a.IsComputed() {
		return false
	}
	switch v := a.(type) {
	case rschema.BoolAttribute:
		return v.Default == nil
	case rschema.StringAttribute:
		return v.Default == nil
	case rschema.Int64Attribute:
		return v.Default == nil
	case rschema.Float64Attribute:
		return v.Default == nil
	case rschema.ListAttribute:
		return v.Default == nil
	}
	return true
}

// nil means the keyed entry has no state yet, which is distinct from a null state attribute.
func stateAttribute(state attr.Value, name string) attr.Value {
	obj, ok := state.(types.Object)
	if !ok || obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	return obj.Attributes()[name]
}

// A null config object leaves every nested attribute null; an unknown one is passed down as is.
func configAttribute(config attr.Value, name string) attr.Value {
	obj, ok := config.(types.Object)
	if !ok || obj.IsUnknown() {
		return config
	}
	if obj.IsNull() {
		return nil
	}
	return obj.Attributes()[name]
}

func configElement(config attr.Value, i int) attr.Value {
	list, ok := config.(types.List)
	if !ok || list.IsUnknown() {
		return config
	}
	if list.IsNull() || i >= len(list.Elements()) {
		return nil
	}
	return list.Elements()[i]
}

func isNull(v attr.Value) bool {
	return v == nil || v.IsNull()
}

func unknownOf(ctx context.Context, v attr.Value) attr.Value {
	t := v.Type(ctx)
	unknown, err := t.ValueFromTerraform(ctx, tftypes.NewValue(t.TerraformType(ctx), tftypes.UnknownValue))
	if err != nil {
		return v
	}
	return unknown
}

func elementKey(e attr.Value, key string) (string, bool) {
	obj, ok := e.(types.Object)
	if !ok || obj.IsNull() || obj.IsUnknown() {
		return "", false
	}

	switch v := obj.Attributes()[key].(type) {
	case types.String:
		if v.IsNull() || v.IsUnknown() {
			return "", false
		}
		return v.ValueString(), true
	case types.Int64:
		if v.IsNull() || v.IsUnknown() {
			return "", false
		}
		return strconv.FormatInt(v.ValueInt64(), 10), true
	}
	return "", false
}

// A plan can still hold unknown nested objects (disk = var.x), which a struct target refuses.
func ObjectsFromList[M any](ctx context.Context, list types.List) ([]M, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diags
	}

	lenient := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	models := make([]M, 0, len(list.Elements()))
	for _, e := range list.Elements() {
		obj, ok := e.(types.Object)
		if !ok || obj.IsNull() || obj.IsUnknown() {
			continue
		}
		var m M
		diags.Append(obj.As(ctx, &m, lenient)...)
		models = append(models, m)
	}

	return models, diags
}

// The API refuses a new user without a password, and silently ignores a password type change without one.
func ValidateClickHouseUserPasswords(root path.Path, prior, planned []ClickHouseClusterModel) diag.Diagnostics {
	var diags diag.Diagnostics

	existing := make(map[string]map[string]types.String, len(prior))
	for _, c := range prior {
		users := make(map[string]types.String, len(c.Users))
		for _, u := range c.Users {
			passwordType := types.StringNull()
			if u.Password != nil {
				passwordType = u.Password.Type
			}
			users[u.Name.ValueString()] = passwordType
		}
		existing[c.Name.ValueString()] = users
	}

	for i, c := range planned {
		if c.Name.IsUnknown() {
			continue
		}
		for j, u := range c.Users {
			if u.Name.IsUnknown() || u.Password == nil || !u.Password.EncValue.IsNull() {
				continue
			}
			encValue := root.AtListIndex(i).AtName("users").AtListIndex(j).AtName("password").AtName("enc_value")
			priorType, exists := existing[c.Name.ValueString()][u.Name.ValueString()]
			switch {
			case !exists:
				diags.AddAttributeError(encValue, "Missing ClickHouse User Password",
					fmt.Sprintf("User %q in cluster %q does not exist yet, and the API requires a password to create it. Set enc_value; it can be omitted later to keep the current password.", u.Name.ValueString(), c.Name.ValueString()))
			case !u.Password.Type.IsUnknown() && !priorType.IsNull() && !u.Password.Type.Equal(priorType):
				diags.AddAttributeError(encValue, "Missing ClickHouse User Password",
					fmt.Sprintf("User %q in cluster %q changes its password type from %s to %s, which needs a new password in that form. Set enc_value.", u.Name.ValueString(), c.Name.ValueString(), priorType.ValueString(), u.Password.Type.ValueString()))
			}
		}
	}

	return diags
}
