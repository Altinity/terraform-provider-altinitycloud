package hosted_env

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func hostedAWSSchema(t *testing.T) tfsdk.Config {
	t.Helper()
	r := &HostedAWSEnvResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	return tfsdk.Config{Schema: resp.Schema}
}

// buildConfig: all attrs null except those in override.
func buildConfig(t *testing.T, cfg tfsdk.Config, override map[string]func(tftypes.Type) tftypes.Value) tftypes.Value {
	t.Helper()
	objType, ok := cfg.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not a tftypes.Object")
	}
	vals := map[string]tftypes.Value{}
	for name, at := range objType.AttributeTypes {
		if fn, ok := override[name]; ok {
			vals[name] = fn(at)
		} else {
			vals[name] = tftypes.NewValue(at, nil)
		}
	}
	return tftypes.NewValue(objType, vals)
}

func runValidate(t *testing.T, raw tftypes.Value, cfg tfsdk.Config) *resource.ValidateConfigResponse {
	t.Helper()
	cfg.Raw = raw
	resp := &resource.ValidateConfigResponse{}
	(&HostedAWSEnvResource{}).ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, resp)
	return resp
}

func unknown(at tftypes.Type) tftypes.Value { return tftypes.NewValue(at, tftypes.UnknownValue) }

func objectWith(t *testing.T, at tftypes.Type, set map[string]any) tftypes.Value {
	t.Helper()
	ot, ok := at.(tftypes.Object)
	if !ok {
		t.Fatal("attribute is not a tftypes.Object")
	}
	fields := map[string]tftypes.Value{}
	for name, ft := range ot.AttributeTypes {
		fields[name] = tftypes.NewValue(ft, set[name])
	}
	return tftypes.NewValue(ot, fields)
}

// Regression: an unknown nested struct-pointer attr must not crash ValidateConfig.
func TestHostedAWSValidateConfigUnknownNestedAttrs(t *testing.T) {
	cfg := hostedAWSSchema(t)

	t.Run("backups unknown, datadog null", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{"backups": unknown})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("every attribute unknown", func(t *testing.T) {
		objType, ok := cfg.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
		if !ok {
			t.Fatal("schema type is not a tftypes.Object")
		}
		all := map[string]func(tftypes.Type) tftypes.Value{}
		for name := range objType.AttributeTypes {
			all[name] = unknown
		}
		raw := buildConfig(t, cfg, all)
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("backups unknown while datadog is a known object", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"backups": unknown,
			"datadog": func(at tftypes.Type) tftypes.Value { return objectWith(t, at, nil) },
		})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	t.Run("iceberg unknown", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{"iceberg": unknown})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	// Sanity: datadog validation still fires through the scoped read.
	t.Run("datadog enabled without api key still errors", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"datadog": func(at tftypes.Type) tftypes.Value {
				return objectWith(t, at, map[string]any{"enabled": true})
			},
		})
		if resp := runValidate(t, raw, cfg); !resp.Diagnostics.HasError() {
			t.Fatal("expected validation error for enabled datadog without enc_api_key")
		}
	})
}
