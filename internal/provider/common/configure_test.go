package common

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Terraform calls Configure before the provider itself is configured, so nil
// provider data must stop the caller without reporting an error.
func TestSDKFromProviderDataNil(t *testing.T) {
	var diags diag.Diagnostics

	got, ok := SDKFromProviderData(nil, &diags)

	if ok || got != nil {
		t.Errorf("got (%v, %v), want (nil, false)", got, ok)
	}
	if diags.HasError() {
		t.Errorf("expected no diagnostics, got: %v", diags.Errors())
	}
}

func TestSDKFromProviderDataWrongType(t *testing.T) {
	var diags diag.Diagnostics

	got, ok := SDKFromProviderData("not the sdk", &diags)

	if ok || got != nil {
		t.Errorf("got (%v, %v), want (nil, false)", got, ok)
	}
	if len(diags.Errors()) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(diags.Errors()), diags.Errors())
	}
	if diags.Errors()[0].Summary() != "Unexpected Resource Configure Type" {
		t.Errorf("unexpected summary: %s", diags.Errors()[0].Summary())
	}
}

func TestSDKFromProviderData(t *testing.T) {
	var diags diag.Diagnostics
	want := &sdk.AltinityCloudSDK{Client: &client.Client{}}

	got, ok := SDKFromProviderData(want, &diags)

	if !ok || got != want {
		t.Errorf("got (%v, %v), want (%v, true)", got, ok, want)
	}
	if diags.HasError() {
		t.Errorf("expected no diagnostics, got: %v", diags.Errors())
	}
}
