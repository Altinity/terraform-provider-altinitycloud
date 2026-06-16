package env

import (
	"context"
	"testing"

	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func customDomainList(vals ...string) types.List {
	elems := make([]attr.Value, len(vals))
	for i, v := range vals {
		elems[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elems)
}

// toModel is state-aware: the field the user manages (per prior state) is refreshed
// from the API, the other stays null. This prevents the API customDomains[0] echo
// from flipping a list-managed env into a permanent diff on the deprecated scalar.
func TestAWSEnvResourceModel_toModel_CustomDomains(t *testing.T) {
	apiSpec := func() *sdk.AWSEnvSpecFragment {
		return &sdk.AWSEnvSpecFragment{
			CustomDomain:          &[]string{"a.com"}[0],
			CustomDomains:         []string{"a.com", "b.com"},
			LoadBalancingStrategy: sdk.LoadBalancingStrategyRoundRobin,
		}
	}

	t.Run("list-managed prior -> refresh list, scalar null", func(t *testing.T) {
		model := &AWSEnvResourceModel{CustomDomains: customDomainList("a.com", "b.com")}
		diags := model.toModel(sdk.GetAWSEnv_AWSEnv{Name: "env", Spec: apiSpec()})
		if diags.HasError() {
			t.Fatalf("toModel: %v", diags)
		}
		if !model.CustomDomain.IsNull() {
			t.Errorf("custom_domain: expected null, got %q", model.CustomDomain.ValueString())
		}
		var got []string
		model.CustomDomains.ElementsAs(context.Background(), &got, false)
		if len(got) != 2 || got[0] != "a.com" || got[1] != "b.com" {
			t.Errorf("custom_domains: expected [a.com b.com], got %v", got)
		}
	})

	t.Run("scalar-managed prior -> mirror scalar, list null", func(t *testing.T) {
		model := &AWSEnvResourceModel{} // prior custom_domains null
		diags := model.toModel(sdk.GetAWSEnv_AWSEnv{Name: "env", Spec: apiSpec()})
		if diags.HasError() {
			t.Fatalf("toModel: %v", diags)
		}
		if model.CustomDomain.ValueString() != "a.com" {
			t.Errorf("custom_domain: expected 'a.com', got %q", model.CustomDomain.ValueString())
		}
		if !model.CustomDomains.IsNull() {
			t.Errorf("custom_domains: expected null, got %v", model.CustomDomains)
		}
	})
}

// toSDK wires the mutually-exclusive fields into both the create and update specs.
func TestAWSEnvResourceModel_toSDK_CustomDomains(t *testing.T) {
	base := func() AWSEnvResourceModel {
		return AWSEnvResourceModel{
			Name:                  types.StringValue("env"),
			Region:                types.StringValue("us-east-1"),
			CIDR:                  types.StringValue("10.0.0.0/16"),
			LoadBalancingStrategy: types.StringValue(string(sdk.LoadBalancingStrategyRoundRobin)),
		}
	}

	t.Run("list set -> spec.CustomDomains, scalar nil", func(t *testing.T) {
		m := base()
		m.CustomDomain = types.StringNull()
		m.CustomDomains = customDomainList("a.com", "b.com")

		create, update, diags := m.toSDK(context.Background())
		if diags.HasError() {
			t.Fatalf("toSDK: %v", diags)
		}
		if create.Spec.CustomDomain != nil {
			t.Errorf("create custom_domain: expected nil, got %q", *create.Spec.CustomDomain)
		}
		if len(create.Spec.CustomDomains) != 2 {
			t.Errorf("create custom_domains: expected 2, got %v", create.Spec.CustomDomains)
		}
		if update.Spec.CustomDomain != nil {
			t.Errorf("update custom_domain: expected nil, got %q", *update.Spec.CustomDomain)
		}
		if len(update.Spec.CustomDomains) != 2 {
			t.Errorf("update custom_domains: expected 2, got %v", update.Spec.CustomDomains)
		}
	})

	t.Run("scalar set -> spec.CustomDomain, list nil", func(t *testing.T) {
		m := base()
		m.CustomDomain = types.StringValue("a.com")
		m.CustomDomains = types.ListNull(types.StringType)

		create, update, diags := m.toSDK(context.Background())
		if diags.HasError() {
			t.Fatalf("toSDK: %v", diags)
		}
		if create.Spec.CustomDomain == nil || *create.Spec.CustomDomain != "a.com" {
			t.Errorf("create custom_domain: expected 'a.com', got %v", create.Spec.CustomDomain)
		}
		if create.Spec.CustomDomains != nil {
			t.Errorf("create custom_domains: expected nil, got %v", create.Spec.CustomDomains)
		}
		if update.Spec.CustomDomain == nil || *update.Spec.CustomDomain != "a.com" {
			t.Errorf("update custom_domain: expected 'a.com', got %v", update.Spec.CustomDomain)
		}
		if update.Spec.CustomDomains != nil {
			t.Errorf("update custom_domains: expected nil, got %v", update.Spec.CustomDomains)
		}
	})
}
