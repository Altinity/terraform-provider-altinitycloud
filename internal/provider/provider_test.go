package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var unknownString = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

func knownString(v string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, v)
}

// configure runs Configure with the given attributes set and every other one null.
// Both env vars are cleared so the ambient shell cannot change the outcome.
func configure(t *testing.T, values map[string]tftypes.Value) *provider.ConfigureResponse {
	t.Helper()
	ctx := context.Background()

	p := New("test")()

	sresp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, sresp)
	if sresp.Diagnostics.HasError() {
		t.Fatalf("schema errored: %v", sresp.Diagnostics.Errors())
	}

	objType, ok := sresp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not a tftypes.Object")
	}

	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		if v, set := values[name]; set {
			vals[name] = v
			continue
		}
		vals[name] = tftypes.NewValue(attrType, nil)
	}

	resp := &provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{
		Config: tfsdk.Config{Schema: sresp.Schema, Raw: tftypes.NewValue(objType, vals)},
	}, resp)

	return resp
}

func assertErrorSummary(t *testing.T, resp *provider.ConfigureResponse, want string) {
	t.Helper()

	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == want {
			return
		}
	}
	t.Fatalf("expected an error with summary %q, got: %v", want, resp.Diagnostics.Errors())
}

func assertNoErrorSummary(t *testing.T, resp *provider.ConfigureResponse, unwanted string) {
	t.Helper()

	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == unwanted {
			t.Fatalf("did not expect an error with summary %q", unwanted)
		}
	}
}

func TestConfigureUnknownAPITokenIsRejected(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"api_token": unknownString})

	assertErrorSummary(t, resp, "Unknown Provider Configuration")
	assertNoErrorSummary(t, resp, "Missing Altinity.Cloud API Token")
}

// An unknown api_token used to be read as "", overwriting the token from the
// environment and failing with "Missing Altinity.Cloud API Token".
func TestConfigureUnknownAPITokenDoesNotFallBackToEnv(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "token-from-env")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"api_token": unknownString})

	assertErrorSummary(t, resp, "Unknown Provider Configuration")
	if resp.ResourceData != nil {
		t.Error("expected the provider not to be configured with the environment token")
	}
}

func TestConfigureUnknownAPIURLIsRejected(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "token")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"api_url": unknownString})

	assertErrorSummary(t, resp, "Unknown Provider Configuration")
}

func TestConfigureUnknownCACrtIsRejected(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "token")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"ca_crt": unknownString})

	assertErrorSummary(t, resp, "Unknown Provider Configuration")
	assertNoErrorSummary(t, resp, "Failed to load CA certificate")
}

func TestConfigureReportsEveryUnknownAttribute(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{
		"api_token": unknownString,
		"api_url":   unknownString,
	})

	if got := len(resp.Diagnostics.Errors()); got != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", got, resp.Diagnostics.Errors())
	}
}

func TestConfigureTokenFromConfig(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"api_token": knownString("token")})

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure errored: %v", resp.Diagnostics.Errors())
	}
	if resp.ResourceData == nil || resp.DataSourceData == nil {
		t.Error("expected the provider to be configured")
	}
}

func TestConfigureTokenFromEnv(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "token-from-env")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, nil)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure errored: %v", resp.Diagnostics.Errors())
	}
	if resp.ResourceData == nil {
		t.Error("expected the provider to be configured")
	}
}

func TestConfigureMissingToken(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, nil)

	assertErrorSummary(t, resp, "Missing Altinity.Cloud API Token")
}

func TestConfigureInvalidCACrt(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "token")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"ca_crt": knownString("not a pem block")})

	assertErrorSummary(t, resp, "Failed to load CA certificate")
}

func TestRequireKnownDetailMentionsEnvVar(t *testing.T) {
	t.Setenv(ENV_VAR_API_TOKEN, "")
	t.Setenv(ENV_VAR_API_URL, "")

	resp := configure(t, map[string]tftypes.Value{"api_token": unknownString})

	errs := resp.Diagnostics.Errors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Detail(), ENV_VAR_API_TOKEN) {
		t.Errorf("detail should name the env var fallback, got: %s", errs[0].Detail())
	}
}
