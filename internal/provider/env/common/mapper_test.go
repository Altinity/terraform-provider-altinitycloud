package env

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCustomDomainsToSDK(t *testing.T) {
	ctx := context.Background()
	list := func(vals ...string) types.List {
		elems := make([]string, len(vals))
		copy(elems, vals)
		l, diags := ListToModel(elems)
		if diags.HasError() {
			t.Fatalf("ListToModel: %v", diags)
		}
		return l
	}
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name          string
		customDomain  types.String
		customDomains types.List
		wantDomain    *string
		wantDomains   []string
		wantErr       bool
	}{
		{
			name:          "list set, scalar null -> list wins, scalar nil",
			customDomain:  types.StringNull(),
			customDomains: list("a.com", "b.com"),
			wantDomain:    nil,
			wantDomains:   []string{"a.com", "b.com"},
		},
		{
			name:          "scalar set, list null -> scalar wins",
			customDomain:  types.StringValue("a.com"),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    strPtr("a.com"),
			wantDomains:   nil,
		},
		{
			name:          "neither set (both null) -> both nil",
			customDomain:  types.StringNull(),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    nil,
			wantDomains:   nil,
		},
		{
			name:          "scalar unknown, list null -> treated as not set",
			customDomain:  types.StringUnknown(),
			customDomains: types.ListNull(types.StringType),
			wantDomain:    nil,
			wantDomains:   nil,
		},
		{
			name:          "list unknown, scalar set -> scalar wins",
			customDomain:  types.StringValue("a.com"),
			customDomains: types.ListUnknown(types.StringType),
			wantDomain:    strPtr("a.com"),
			wantDomains:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, gotDomains, diags := CustomDomainsToSDK(ctx, tt.customDomain, tt.customDomains)
			if diags.HasError() != tt.wantErr {
				t.Fatalf("diags error = %v, want %v (%v)", diags.HasError(), tt.wantErr, diags)
			}
			if (gotDomain == nil) != (tt.wantDomain == nil) {
				t.Fatalf("domain nil mismatch: got %v, want %v", gotDomain, tt.wantDomain)
			}
			if gotDomain != nil && *gotDomain != *tt.wantDomain {
				t.Fatalf("domain = %q, want %q", *gotDomain, *tt.wantDomain)
			}
			if len(gotDomains) != len(tt.wantDomains) {
				t.Fatalf("domains = %v, want %v", gotDomains, tt.wantDomains)
			}
			for i := range gotDomains {
				if gotDomains[i] != tt.wantDomains[i] {
					t.Fatalf("domains[%d] = %q, want %q", i, gotDomains[i], tt.wantDomains[i])
				}
			}
		})
	}
}

func TestCustomDomainsToModel(t *testing.T) {
	list := func(vals ...string) types.List {
		l, diags := ListToModel(vals)
		if diags.HasError() {
			t.Fatalf("ListToModel: %v", diags)
		}
		return l
	}
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		prior          types.List
		specDomain     *string
		specDomains    []string
		wantDomainNull bool
		wantDomain     string
		wantListNull   bool
		wantList       []string
	}{
		{
			name:           "list-managed (prior list set) -> refresh list, scalar null",
			prior:          list("a.com", "b.com"),
			specDomain:     strPtr("a.com"),
			specDomains:    []string{"a.com", "b.com"},
			wantDomainNull: true,
			wantListNull:   false,
			wantList:       []string{"a.com", "b.com"},
		},
		{
			name:           "scalar-managed (prior list null) -> mirror scalar, list null",
			prior:          types.ListNull(types.StringType),
			specDomain:     strPtr("a.com"),
			specDomains:    []string{"a.com"},
			wantDomainNull: false,
			wantDomain:     "a.com",
			wantListNull:   true,
		},
		{
			name:           "neither set -> both null",
			prior:          types.ListNull(types.StringType),
			specDomain:     nil,
			specDomains:    nil,
			wantDomainNull: true,
			wantListNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDomain, gotList, diags := CustomDomainsToModel(tt.prior, tt.specDomain, tt.specDomains)
			if diags.HasError() {
				t.Fatalf("unexpected diags: %v", diags)
			}
			if gotDomain.IsNull() != tt.wantDomainNull {
				t.Fatalf("domain null = %v, want %v (val %q)", gotDomain.IsNull(), tt.wantDomainNull, gotDomain.ValueString())
			}
			if !tt.wantDomainNull && gotDomain.ValueString() != tt.wantDomain {
				t.Fatalf("domain = %q, want %q", gotDomain.ValueString(), tt.wantDomain)
			}
			if gotList.IsNull() != tt.wantListNull {
				t.Fatalf("list null = %v, want %v", gotList.IsNull(), tt.wantListNull)
			}
			if !tt.wantListNull {
				var got []string
				gotList.ElementsAs(context.Background(), &got, false)
				if len(got) != len(tt.wantList) {
					t.Fatalf("list = %v, want %v", got, tt.wantList)
				}
				for i := range got {
					if got[i] != tt.wantList[i] {
						t.Fatalf("list[%d] = %q, want %q", i, got[i], tt.wantList[i])
					}
				}
			}
		})
	}
}
