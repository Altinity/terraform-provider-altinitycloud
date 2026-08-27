package env

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func awsSchema(t *testing.T) tfsdk.Config {
	t.Helper()
	r := &AWSEnvResource{}
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
	(&AWSEnvResource{}).ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, resp)
	return resp
}

func unknown(at tftypes.Type) tftypes.Value { return tftypes.NewValue(at, tftypes.UnknownValue) }

// Regression: an unknown nested struct-pointer attr must not crash ValidateConfig.
func TestAWSValidateConfigUnknownNestedAttrs(t *testing.T) {
	cfg := awsSchema(t)

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
		// known datadog (enabled null) exercises the second read with backups unknown.
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"backups": unknown,
			"datadog": func(at tftypes.Type) tftypes.Value {
				ot, ok := at.(tftypes.Object)
				if !ok {
					t.Fatal("datadog attr is not a tftypes.Object")
				}
				fields := map[string]tftypes.Value{}
				for n, ft := range ot.AttributeTypes {
					fields[n] = tftypes.NewValue(ft, nil)
				}
				return tftypes.NewValue(ot, fields)
			},
		})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	// Sanity: datadog validation still fires through the scoped read.
	t.Run("datadog enabled without api key still errors", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"datadog": func(at tftypes.Type) tftypes.Value {
				ot, ok := at.(tftypes.Object)
				if !ok {
					t.Fatal("datadog attr is not a tftypes.Object")
				}
				fields := map[string]tftypes.Value{}
				for n, ft := range ot.AttributeTypes {
					switch n {
					case "enabled":
						fields[n] = tftypes.NewValue(ft, true)
					case "enc_api_key":
						fields[n] = tftypes.NewValue(ft, nil)
					default:
						fields[n] = tftypes.NewValue(ft, nil)
					}
				}
				return tftypes.NewValue(ot, fields)
			},
		})
		if resp := runValidate(t, raw, cfg); !resp.Diagnostics.HasError() {
			t.Fatal("expected validation error for enabled datadog without enc_api_key")
		}
	})
}

func nullObject(t *testing.T, at tftypes.Type, override map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	ot, ok := at.(tftypes.Object)
	if !ok {
		t.Fatalf("expected a tftypes.Object, got %T", at)
	}

	fields := map[string]tftypes.Value{}
	for name, ft := range ot.AttributeTypes {
		if v, ok := override[name]; ok {
			fields[name] = v
		} else {
			fields[name] = tftypes.NewValue(ft, nil)
		}
	}

	return tftypes.NewValue(ot, fields)
}

// Sanity: ClickHouse validation still fires through the scoped read.
func TestAWSValidateConfigClickHouse(t *testing.T) {
	cfg := awsSchema(t)

	clusters := func(keeperName string) func(tftypes.Type) tftypes.Value {
		return func(at tftypes.Type) tftypes.Value {
			lt, ok := at.(tftypes.List)
			if !ok {
				t.Fatalf("clickhouse_clusters is not a tftypes.List, got %T", at)
			}
			ot, ok := lt.ElementType.(tftypes.Object)
			if !ok {
				t.Fatalf("cluster element is not a tftypes.Object, got %T", lt.ElementType)
			}

			cluster := nullObject(t, ot, map[string]tftypes.Value{
				"name": tftypes.NewValue(ot.AttributeTypes["name"], "ch"),
				"keeper": nullObject(t, ot.AttributeTypes["keeper"], map[string]tftypes.Value{
					"enabled": tftypes.NewValue(tftypes.Bool, true),
					"name":    tftypes.NewValue(tftypes.String, keeperName),
				}),
			})

			return tftypes.NewValue(lt, []tftypes.Value{cluster})
		}
	}

	t.Run("a cluster referencing an undeclared keeper errors", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"clickhouse_clusters": clusters("missing"),
		})
		if resp := runValidate(t, raw, cfg); !resp.Diagnostics.HasError() {
			t.Fatal("expected an error for a keeper that is not declared")
		}
	})

	t.Run("unknown clusters are skipped rather than rejected", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"clickhouse_clusters": unknown,
		})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})

	// Reflecting an unknown into *ClickHouseDiskModel is a hard provider error, and
	// a nested attribute can be unknown while the element around it is not.
	t.Run("an unknown nested attribute defers validation", func(t *testing.T) {
		raw := buildConfig(t, cfg, map[string]func(tftypes.Type) tftypes.Value{
			"clickhouse_clusters": func(at tftypes.Type) tftypes.Value {
				lt, ok := at.(tftypes.List)
				if !ok {
					t.Fatalf("clickhouse_clusters is not a tftypes.List, got %T", at)
				}
				ot, ok := lt.ElementType.(tftypes.Object)
				if !ok {
					t.Fatalf("cluster element is not a tftypes.Object, got %T", lt.ElementType)
				}

				cluster := nullObject(t, ot, map[string]tftypes.Value{
					"name": tftypes.NewValue(ot.AttributeTypes["name"], "ch"),
					"disk": tftypes.NewValue(ot.AttributeTypes["disk"], tftypes.UnknownValue),
				})
				return tftypes.NewValue(lt, []tftypes.Value{cluster})
			},
		})
		if resp := runValidate(t, raw, cfg); resp.Diagnostics.HasError() {
			t.Fatalf("errored: %v", resp.Diagnostics.Errors())
		}
	})
}
