package env

import (
	"context"
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	sdk "github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReorderNodeGroups(t *testing.T) {
	tests := []struct {
		name           string
		model          []common.NodeGroupsModel
		apiNodeGroups  []*sdk.AWSEnvSpecFragment_NodeGroups
		expectedOrder  []string
		expectedLength int
		validateData   bool
	}{
		{
			name: "Preserve model order and add new API node groups",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
			},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerZone: 1},
			},
			expectedOrder:  []string{"system", "user", "monitoring"},
			expectedLength: 3,
		},
		{
			name: "All API node groups exist in model",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
			},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name: "Model has more node groups than API",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
				{NodeType: types.StringValue("missing")},
			},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:  "Empty model with API node groups",
			model: []common.NodeGroupsModel{},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:           "Empty inputs",
			model:          []common.NodeGroupsModel{},
			apiNodeGroups:  []*sdk.AWSEnvSpecFragment_NodeGroups{},
			expectedOrder:  []string{},
			expectedLength: 0,
		},
		{
			name: "Multiple new API node groups",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
			},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerZone: 1},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
				{NodeType: "logging", Name: "logging-group", CapacityPerZone: 1},
				{NodeType: "metrics", Name: "metrics-group", CapacityPerZone: 1},
			},
			expectedOrder:  []string{"system", "monitoring", "logging", "metrics"},
			expectedLength: 4,
		},
		{
			name: "No data loss validation",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
			},
			apiNodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerZone: 5},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 10},
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerZone: 3},
			},
			expectedOrder:  []string{"system", "user", "monitoring"},
			expectedLength: 3,
			validateData:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reorderNodeGroups(tt.model, tt.apiNodeGroups)

			if len(result) != tt.expectedLength {
				t.Errorf("Expected length %d, got %d", tt.expectedLength, len(result))
			}

			for i, expected := range tt.expectedOrder {
				if i >= len(result) {
					t.Errorf("Result has fewer elements than expected")
					break
				}
				if result[i].NodeType != expected {
					t.Errorf("Node type at position %d: expected %s, got %s", i, expected, result[i].NodeType)
				}
			}

			if tt.validateData {

				if result[0].NodeType != "system" {
					t.Errorf("Expected first node type to be 'system', got '%s'", result[0].NodeType)
				}
				if result[0].Name != "system-group" {
					t.Errorf("Expected system Name to be 'system-group', got '%s'", result[0].Name)
				}
				if result[0].CapacityPerZone != 10 {
					t.Errorf("Expected system CapacityPerZone to be 10, got %d", result[0].CapacityPerZone)
				}

				nodeTypeToGroup := make(map[string]*sdk.AWSEnvSpecFragment_NodeGroups)
				for _, group := range result {
					nodeTypeToGroup[group.NodeType] = group
				}

				if nodeTypeToGroup["user"].Name != "user-group" {
					t.Errorf("Expected user Name to be 'user-group', got '%s'", nodeTypeToGroup["user"].Name)
				}
				if nodeTypeToGroup["user"].CapacityPerZone != 5 {
					t.Errorf("Expected user CapacityPerZone to be 5, got %d", nodeTypeToGroup["user"].CapacityPerZone)
				}

				if nodeTypeToGroup["monitoring"].Name != "monitoring-group" {
					t.Errorf("Expected monitoring Name to be 'monitoring-group', got '%s'", nodeTypeToGroup["monitoring"].Name)
				}
				if nodeTypeToGroup["monitoring"].CapacityPerZone != 3 {
					t.Errorf("Expected monitoring CapacityPerZone to be 3, got %d", nodeTypeToGroup["monitoring"].CapacityPerZone)
				}
			}
		})
	}
}

func TestMaintenanceWindowsToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    []*sdk.AWSEnvSpecFragment_MaintenanceWindows
		expected []struct {
			name          string
			hour          int64
			lengthInHours int64
			days          []string
		}
		expectEmpty bool
	}{
		{
			name: "Multiple maintenance windows",
			input: []*sdk.AWSEnvSpecFragment_MaintenanceWindows{
				{
					Name:          "weekly-maintenance",
					Hour:          2,
					LengthInHours: 4,
					Days:          []sdk.Day{"saturday", "sunday"},
				},
				{
					Name:          "daily-maintenance",
					Hour:          1,
					LengthInHours: 1,
					Days:          []sdk.Day{"monday", "tuesday", "wednesday"},
				},
			},
			expected: []struct {
				name          string
				hour          int64
				lengthInHours int64
				days          []string
			}{
				{
					name:          "weekly-maintenance",
					hour:          2,
					lengthInHours: 4,
					days:          []string{"saturday", "sunday"},
				},
				{
					name:          "daily-maintenance",
					hour:          1,
					lengthInHours: 1,
					days:          []string{"monday", "tuesday", "wednesday"},
				},
			},
		},
		{
			name:        "Nil input",
			input:       nil,
			expectEmpty: true,
		},
		{
			name:        "Empty slice input",
			input:       []*sdk.AWSEnvSpecFragment_MaintenanceWindows{},
			expectEmpty: true,
		},
		{
			name: "Single maintenance window",
			input: []*sdk.AWSEnvSpecFragment_MaintenanceWindows{
				{
					Name:          "nightly-backup",
					Hour:          3,
					LengthInHours: 2,
					Days:          []sdk.Day{"friday"},
				},
			},
			expected: []struct {
				name          string
				hour          int64
				lengthInHours int64
				days          []string
			}{
				{
					name:          "nightly-backup",
					hour:          3,
					lengthInHours: 2,
					days:          []string{"friday"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maintenanceWindowsToModel(tt.input)

			if tt.expectEmpty {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %d items", len(result))
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d maintenance windows, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].Name.ValueString() != expected.name {
					t.Errorf("Window %d name: expected '%s', got '%s'", i, expected.name, result[i].Name.ValueString())
				}
				if result[i].Hour.ValueInt64() != expected.hour {
					t.Errorf("Window %d hour: expected %d, got %d", i, expected.hour, result[i].Hour.ValueInt64())
				}
				if result[i].LengthInHours.ValueInt64() != expected.lengthInHours {
					t.Errorf("Window %d length: expected %d, got %d", i, expected.lengthInHours, result[i].LengthInHours.ValueInt64())
				}
				if len(result[i].Days) != len(expected.days) {
					t.Errorf("Window %d days count: expected %d, got %d", i, len(expected.days), len(result[i].Days))
				} else {
					for j, expectedDay := range expected.days {
						if result[i].Days[j].ValueString() != expectedDay {
							t.Errorf("Window %d day %d: expected '%s', got '%s'", i, j, expectedDay, result[i].Days[j].ValueString())
						}
					}
				}
			}
		})
	}
}

func TestLoadBalancersToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    *LoadBalancersModel
		expected *sdk.AWSEnvLoadBalancersSpecInput
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Complete load balancers config",
			input: &LoadBalancersModel{
				Public: &PublicLoadBalancerModel{
					Enabled:        types.BoolValue(true),
					SourceIPRanges: []types.String{types.StringValue("0.0.0.0/0"), types.StringValue("192.168.1.0/24")},
					CrossZone:      types.BoolValue(false),
				},
				Internal: &InternalLoadBalancerModel{
					Enabled:                          types.BoolValue(true),
					SourceIPRanges:                   []types.String{types.StringValue("10.0.0.0/8")},
					CrossZone:                        types.BoolValue(true),
					EndpointServiceAllowedPrincipals: []types.String{types.StringValue("arn:aws:iam::123456789012:root")},
				},
			},
			expected: &sdk.AWSEnvLoadBalancersSpecInput{
				Public: &sdk.AWSEnvLoadBalancerPublicSpecInput{
					Enabled:        &[]bool{true}[0],
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
					CrossZone:      &[]bool{false}[0],
				},
				Internal: &sdk.AWSEnvLoadBalancerInternalSpecInput{
					Enabled:                          &[]bool{true}[0],
					SourceIPRanges:                   []string{"10.0.0.0/8"},
					CrossZone:                        &[]bool{true}[0],
					EndpointServiceAllowedPrincipals: []string{"arn:aws:iam::123456789012:root"},
				},
			},
		},
		{
			name: "Only public load balancer",
			input: &LoadBalancersModel{
				Public: &PublicLoadBalancerModel{
					Enabled:   types.BoolValue(false),
					CrossZone: types.BoolValue(true),
				},
			},
			expected: &sdk.AWSEnvLoadBalancersSpecInput{
				Public: &sdk.AWSEnvLoadBalancerPublicSpecInput{
					Enabled:        &[]bool{false}[0],
					SourceIPRanges: []string{},
					CrossZone:      &[]bool{true}[0],
				},
			},
		},
		{
			name: "Only internal load balancer",
			input: &LoadBalancersModel{
				Internal: &InternalLoadBalancerModel{
					Enabled:   types.BoolValue(true),
					CrossZone: types.BoolValue(false),
				},
			},
			expected: &sdk.AWSEnvLoadBalancersSpecInput{
				Internal: &sdk.AWSEnvLoadBalancerInternalSpecInput{
					Enabled:                          &[]bool{true}[0],
					SourceIPRanges:                   []string{},
					CrossZone:                        &[]bool{false}[0],
					EndpointServiceAllowedPrincipals: []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loadBalancersToSDK(tt.input)

			if (tt.expected == nil) != (result == nil) {
				t.Errorf("Expected nil: %v, got nil: %v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected != nil && result != nil {

				if (tt.expected.Public == nil) != (result.Public == nil) {
					t.Errorf("Public load balancer presence mismatch")
				}
				if tt.expected.Public != nil && result.Public != nil {
					if *tt.expected.Public.Enabled != *result.Public.Enabled {
						t.Errorf("Public enabled mismatch: expected %v, got %v", *tt.expected.Public.Enabled, *result.Public.Enabled)
					}
					if *tt.expected.Public.CrossZone != *result.Public.CrossZone {
						t.Errorf("Public CrossZone mismatch: expected %v, got %v", *tt.expected.Public.CrossZone, *result.Public.CrossZone)
					}
					if len(tt.expected.Public.SourceIPRanges) != len(result.Public.SourceIPRanges) {
						t.Errorf("Public SourceIPRanges length mismatch: expected %d, got %d", len(tt.expected.Public.SourceIPRanges), len(result.Public.SourceIPRanges))
					}
				}

				if (tt.expected.Internal == nil) != (result.Internal == nil) {
					t.Errorf("Internal load balancer presence mismatch")
				}
				if tt.expected.Internal != nil && result.Internal != nil {
					if *tt.expected.Internal.Enabled != *result.Internal.Enabled {
						t.Errorf("Internal enabled mismatch: expected %v, got %v", *tt.expected.Internal.Enabled, *result.Internal.Enabled)
					}
					if *tt.expected.Internal.CrossZone != *result.Internal.CrossZone {
						t.Errorf("Internal CrossZone mismatch: expected %v, got %v", *tt.expected.Internal.CrossZone, *result.Internal.CrossZone)
					}
				}
			}
		})
	}
}

func TestLoadBalancersToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.AWSEnvSpecFragment_LoadBalancers
		expected struct {
			publicEnabled           bool
			publicCrossZone         bool
			publicSourceIPCount     int
			internalEnabled         bool
			internalCrossZone       bool
			internalSourceIPCount   int
			endpointPrincipalsCount int
		}
	}{
		{
			name: "Complete load balancers",
			input: sdk.AWSEnvSpecFragment_LoadBalancers{
				Public: sdk.AWSEnvSpecFragment_LoadBalancers_Public{
					Enabled:        true,
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
					CrossZone:      false,
				},
				Internal: sdk.AWSEnvSpecFragment_LoadBalancers_Internal{
					Enabled:                          true,
					SourceIPRanges:                   []string{"10.0.0.0/8"},
					CrossZone:                        true,
					EndpointServiceAllowedPrincipals: []string{"arn:aws:iam::123456789012:root", "arn:aws:iam::987654321098:root"},
				},
			},
			expected: struct {
				publicEnabled           bool
				publicCrossZone         bool
				publicSourceIPCount     int
				internalEnabled         bool
				internalCrossZone       bool
				internalSourceIPCount   int
				endpointPrincipalsCount int
			}{
				publicEnabled:           true,
				publicCrossZone:         false,
				publicSourceIPCount:     2,
				internalEnabled:         true,
				internalCrossZone:       true,
				internalSourceIPCount:   1,
				endpointPrincipalsCount: 2,
			},
		},
		{
			name: "Minimal load balancers",
			input: sdk.AWSEnvSpecFragment_LoadBalancers{
				Public: sdk.AWSEnvSpecFragment_LoadBalancers_Public{
					Enabled:        false,
					SourceIPRanges: []string{},
					CrossZone:      true,
				},
				Internal: sdk.AWSEnvSpecFragment_LoadBalancers_Internal{
					Enabled:                          false,
					SourceIPRanges:                   []string{},
					CrossZone:                        false,
					EndpointServiceAllowedPrincipals: []string{},
				},
			},
			expected: struct {
				publicEnabled           bool
				publicCrossZone         bool
				publicSourceIPCount     int
				internalEnabled         bool
				internalCrossZone       bool
				internalSourceIPCount   int
				endpointPrincipalsCount int
			}{
				publicEnabled:           false,
				publicCrossZone:         true,
				publicSourceIPCount:     0,
				internalEnabled:         false,
				internalCrossZone:       false,
				internalSourceIPCount:   0,
				endpointPrincipalsCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loadBalancersToModel(tt.input)

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			if result.Public == nil {
				t.Error("Expected non-nil Public load balancer")
				return
			}
			if result.Public.Enabled.ValueBool() != tt.expected.publicEnabled {
				t.Errorf("Public enabled: expected %v, got %v", tt.expected.publicEnabled, result.Public.Enabled.ValueBool())
			}
			if result.Public.CrossZone.ValueBool() != tt.expected.publicCrossZone {
				t.Errorf("Public CrossZone: expected %v, got %v", tt.expected.publicCrossZone, result.Public.CrossZone.ValueBool())
			}
			if len(result.Public.SourceIPRanges) != tt.expected.publicSourceIPCount {
				t.Errorf("Public SourceIPRanges count: expected %d, got %d", tt.expected.publicSourceIPCount, len(result.Public.SourceIPRanges))
			}

			if result.Internal == nil {
				t.Error("Expected non-nil Internal load balancer")
				return
			}
			if result.Internal.Enabled.ValueBool() != tt.expected.internalEnabled {
				t.Errorf("Internal enabled: expected %v, got %v", tt.expected.internalEnabled, result.Internal.Enabled.ValueBool())
			}
			if result.Internal.CrossZone.ValueBool() != tt.expected.internalCrossZone {
				t.Errorf("Internal CrossZone: expected %v, got %v", tt.expected.internalCrossZone, result.Internal.CrossZone.ValueBool())
			}
			if len(result.Internal.SourceIPRanges) != tt.expected.internalSourceIPCount {
				t.Errorf("Internal SourceIPRanges count: expected %d, got %d", tt.expected.internalSourceIPCount, len(result.Internal.SourceIPRanges))
			}
			if len(result.Internal.EndpointServiceAllowedPrincipals) != tt.expected.endpointPrincipalsCount {
				t.Errorf("EndpointServiceAllowedPrincipals count: expected %d, got %d", tt.expected.endpointPrincipalsCount, len(result.Internal.EndpointServiceAllowedPrincipals))
			}
		})
	}
}

func TestNodeGroupsToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    []common.NodeGroupsModel
		expected []struct {
			name            string
			nodeType        string
			capacityPerZone int64
			zonesCount      int
		}
	}{
		{
			name: "Multiple node groups",
			input: []common.NodeGroupsModel{
				{
					Name:            types.StringValue("system-group"),
					NodeType:        types.StringValue("system"),
					CapacityPerZone: types.Int64Value(2),
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1a"), types.StringValue("us-east-1b")}),
					Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
				{
					Name:            types.StringValue("user-group"),
					NodeType:        types.StringValue("user"),
					CapacityPerZone: types.Int64Value(5),
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1c")}),
					Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
			},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{
				{
					name:            "system-group",
					nodeType:        "system",
					capacityPerZone: 2,
					zonesCount:      2,
				},
				{
					name:            "user-group",
					nodeType:        "user",
					capacityPerZone: 5,
					zonesCount:      1,
				},
			},
		},
		{
			name:  "Empty input",
			input: []common.NodeGroupsModel{},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{},
		},
		{
			name: "Single node group",
			input: []common.NodeGroupsModel{
				{
					Name:            types.StringValue("monitoring"),
					NodeType:        types.StringValue("monitoring"),
					CapacityPerZone: types.Int64Value(1),
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1a")}),
					Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
			},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{
				{
					name:            "monitoring",
					nodeType:        "monitoring",
					capacityPerZone: 1,
					zonesCount:      1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nodeGroupsToSDK(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d node groups, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].NodeType != expected.nodeType {
					t.Errorf("Node group %d NodeType: expected '%s', got '%s'", i, expected.nodeType, result[i].NodeType)
				}
				if *result[i].Name != expected.name {
					t.Errorf("Node group %d Name: expected '%s', got '%s'", i, expected.name, *result[i].Name)
				}
				if result[i].CapacityPerZone != expected.capacityPerZone {
					t.Errorf("Node group %d CapacityPerZone: expected %d, got %d", i, expected.capacityPerZone, result[i].CapacityPerZone)
				}
				if len(result[i].Zones) != expected.zonesCount {
					t.Errorf("Node group %d Zones count: expected %d, got %d", i, expected.zonesCount, len(result[i].Zones))
				}
			}
		})
	}
}

func TestNodeGroupsToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    []*sdk.AWSEnvSpecFragment_NodeGroups
		expected []struct {
			name            string
			nodeType        string
			capacityPerZone int64
			zonesCount      int
		}
	}{
		{
			name: "Multiple node groups",
			input: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{
					Name:            "system-group",
					NodeType:        "system",
					CapacityPerZone: 3,
					Zones:           []string{"us-east-1a", "us-east-1b", "us-east-1c"},
				},
				{
					Name:            "user-group",
					NodeType:        "user",
					CapacityPerZone: 10,
					Zones:           []string{"us-east-1a"},
				},
			},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{
				{
					name:            "system-group",
					nodeType:        "system",
					capacityPerZone: 3,
					zonesCount:      3,
				},
				{
					name:            "user-group",
					nodeType:        "user",
					capacityPerZone: 10,
					zonesCount:      1,
				},
			},
		},
		{
			name:  "Empty input",
			input: []*sdk.AWSEnvSpecFragment_NodeGroups{},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{},
		},
		{
			name: "Single node group",
			input: []*sdk.AWSEnvSpecFragment_NodeGroups{
				{
					Name:            "logging",
					NodeType:        "logging",
					CapacityPerZone: 2,
					Zones:           []string{"us-east-1a", "us-east-1b"},
				},
			},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{
				{
					name:            "logging",
					nodeType:        "logging",
					capacityPerZone: 2,
					zonesCount:      2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nodeGroupsToModel(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d node groups, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].NodeType.ValueString() != expected.nodeType {
					t.Errorf("Node group %d NodeType: expected '%s', got '%s'", i, expected.nodeType, result[i].NodeType.ValueString())
				}
				if result[i].Name.ValueString() != expected.name {
					t.Errorf("Node group %d Name: expected '%s', got '%s'", i, expected.name, result[i].Name.ValueString())
				}
				if result[i].CapacityPerZone.ValueInt64() != expected.capacityPerZone {
					t.Errorf("Node group %d CapacityPerZone: expected %d, got %d", i, expected.capacityPerZone, result[i].CapacityPerZone.ValueInt64())
				}

				var zones []string
				result[i].Zones.ElementsAs(context.TODO(), &zones, false)
				if len(zones) != expected.zonesCount {
					t.Errorf("Node group %d Zones count: expected %d, got %d", i, expected.zonesCount, len(zones))
				}
			}
		})
	}
}

func TestAWSEnvResourceModel_toSDK(t *testing.T) {
	tests := []struct {
		name     string
		model    AWSEnvResourceModel
		validate func(t *testing.T, create sdk.CreateAWSEnvInput, update sdk.UpdateAWSEnvInput)
	}{
		{
			name: "Complete model with all fields",
			model: AWSEnvResourceModel{
				Name:                         types.StringValue("test-env"),
				CustomDomain:                 types.StringValue("custom.example.com"),
				LoadBalancingStrategy:        types.StringValue("round_robin"),
				Region:                       types.StringValue("us-east-1"),
				PermissionsBoundaryPolicyArn: types.StringValue("arn:aws:iam::123456789012:policy/boundary"),
				ResourcePrefix:               types.StringValue("altinity-"),
				NAT:                          types.BoolValue(true),
				CIDR:                         types.StringValue("10.0.0.0/16"),
				AWSAccountID:                 types.StringValue("123456789012"),
				Zones:                        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1a"), types.StringValue("us-east-1b")}),
				LoadBalancers: &LoadBalancersModel{
					Public: &PublicLoadBalancerModel{
						Enabled:        types.BoolValue(true),
						SourceIPRanges: []types.String{types.StringValue("0.0.0.0/0")},
						CrossZone:      types.BoolValue(false),
					},
				},
				NodeGroups: []common.NodeGroupsModel{
					{
						Name:            types.StringValue("system"),
						NodeType:        types.StringValue("system"),
						CapacityPerZone: types.Int64Value(2),
						Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-east-1a")}),
						Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
					},
				},
				PeeringConnections: []AWSEnvPeeringConnectionModel{
					{
						AWSAccountID: types.StringValue("987654321098"),
						VpcID:        types.StringValue("vpc-12345"),
						VpcRegion:    types.StringValue("us-west-2"),
					},
				},
				Endpoints: []AWSEnvEndpointModel{
					{
						ServiceName: types.StringValue("com.amazonaws.vpce.us-east-1.s3"),
						Alias:       types.StringValue("s3-alias"),
						PrivateDNS:  types.BoolValue(true),
					},
				},
				Tags: []common.KeyValueModel{
					{
						Key:   types.StringValue("Environment"),
						Value: types.StringValue("test"),
					},
				},
				CloudConnect: types.BoolValue(true),
				MaintenanceWindows: []common.MaintenanceWindowModel{
					{
						Name:          types.StringValue("weekly"),
						Enabled:       types.BoolValue(true),
						Hour:          types.Int64Value(2),
						LengthInHours: types.Int64Value(4),
						Days:          []types.String{types.StringValue("saturday")},
					},
				},
			},
			validate: func(t *testing.T, create sdk.CreateAWSEnvInput, update sdk.UpdateAWSEnvInput) {

				if create.Name != "test-env" {
					t.Errorf("Create name: expected 'test-env', got '%s'", create.Name)
				}
				if create.Spec == nil {
					t.Fatal("Create spec should not be nil")
				}
				if *create.Spec.CustomDomain != "custom.example.com" {
					t.Errorf("Create custom domain: expected 'custom.example.com', got '%s'", *create.Spec.CustomDomain)
				}
				if create.Spec.Region != "us-east-1" {
					t.Errorf("Create region: expected 'us-east-1', got '%s'", create.Spec.Region)
				}
				if create.Spec.AWSAccountID != "123456789012" {
					t.Errorf("Create AWS account ID: expected '123456789012', got '%s'", create.Spec.AWSAccountID)
				}
				if create.Spec.Cidr != "10.0.0.0/16" {
					t.Errorf("Create CIDR: expected '10.0.0.0/16', got '%s'", create.Spec.Cidr)
				}
				if len(create.Spec.Zones) != 2 {
					t.Errorf("Create zones: expected 2, got %d", len(create.Spec.Zones))
				}
				if len(create.Spec.NodeGroups) != 1 {
					t.Errorf("Create node groups: expected 1, got %d", len(create.Spec.NodeGroups))
				}
				if len(create.Spec.PeeringConnections) != 1 {
					t.Errorf("Create peering connections: expected 1, got %d", len(create.Spec.PeeringConnections))
				}
				if len(create.Spec.Endpoints) != 1 {
					t.Errorf("Create endpoints: expected 1, got %d", len(create.Spec.Endpoints))
				}
				if len(create.Spec.Tags) != 1 {
					t.Errorf("Create tags: expected 1, got %d", len(create.Spec.Tags))
				}

				if update.Name != "test-env" {
					t.Errorf("Update name: expected 'test-env', got '%s'", update.Name)
				}
				if update.Spec == nil {
					t.Fatal("Update spec should not be nil")
				}
				if *update.Spec.CustomDomain != "custom.example.com" {
					t.Errorf("Update custom domain: expected 'custom.example.com', got '%s'", *update.Spec.CustomDomain)
				}
			},
		},
		{
			name: "Minimal model with required fields only",
			model: AWSEnvResourceModel{
				Name:         types.StringValue("minimal-env"),
				Region:       types.StringValue("us-west-2"),
				CIDR:         types.StringValue("172.16.0.0/16"),
				AWSAccountID: types.StringValue("111122223333"),
				Zones:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("us-west-2a")}),
				NodeGroups:   []common.NodeGroupsModel{},
				CloudConnect: types.BoolValue(false),
			},
			validate: func(t *testing.T, create sdk.CreateAWSEnvInput, update sdk.UpdateAWSEnvInput) {
				if create.Name != "minimal-env" {
					t.Errorf("Create name: expected 'minimal-env', got '%s'", create.Name)
				}
				if create.Spec.Region != "us-west-2" {
					t.Errorf("Create region: expected 'us-west-2', got '%s'", create.Spec.Region)
				}
				if create.Spec.Cidr != "172.16.0.0/16" {
					t.Errorf("Create CIDR: expected '172.16.0.0/16', got '%s'", create.Spec.Cidr)
				}
				if create.Spec.AWSAccountID != "111122223333" {
					t.Errorf("Create AWS account ID: expected '111122223333', got '%s'", create.Spec.AWSAccountID)
				}
				if len(create.Spec.NodeGroups) != 0 {
					t.Errorf("Create node groups: expected 0, got %d", len(create.Spec.NodeGroups))
				}
				if *create.Spec.CloudConnect != false {
					t.Errorf("Create cloud connect: expected false, got %v", *create.Spec.CloudConnect)
				}
			},
		},
		{
			name: "Model with empty optional slices",
			model: AWSEnvResourceModel{
				Name:               types.StringValue("empty-slices"),
				Region:             types.StringValue("eu-west-1"),
				CIDR:               types.StringValue("192.168.0.0/16"),
				AWSAccountID:       types.StringValue("444455556666"),
				Zones:              types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eu-west-1a")}),
				NodeGroups:         []common.NodeGroupsModel{},
				PeeringConnections: []AWSEnvPeeringConnectionModel{},
				Endpoints:          []AWSEnvEndpointModel{},
				Tags:               []common.KeyValueModel{},
				MaintenanceWindows: []common.MaintenanceWindowModel{},
				CloudConnect:       types.BoolValue(true),
			},
			validate: func(t *testing.T, create sdk.CreateAWSEnvInput, update sdk.UpdateAWSEnvInput) {
				if len(create.Spec.NodeGroups) != 0 {
					t.Errorf("Expected empty node groups, got %d", len(create.Spec.NodeGroups))
				}
				if len(create.Spec.PeeringConnections) != 0 {
					t.Errorf("Expected empty peering connections, got %d", len(create.Spec.PeeringConnections))
				}
				if len(create.Spec.Endpoints) != 0 {
					t.Errorf("Expected empty endpoints, got %d", len(create.Spec.Endpoints))
				}
				if len(create.Spec.Tags) != 0 {
					t.Errorf("Expected empty tags, got %d", len(create.Spec.Tags))
				}
				if len(create.Spec.MaintenanceWindows) != 0 {
					t.Errorf("Expected empty maintenance windows, got %d", len(create.Spec.MaintenanceWindows))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			create, update := tt.model.toSDK()
			tt.validate(t, create, update)
		})
	}
}

func TestAWSEnvResourceModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    sdk.GetAWSEnv_AWSEnv
		validate func(t *testing.T, model *AWSEnvResourceModel)
	}{
		{
			name: "Complete SDK response with all fields",
			input: sdk.GetAWSEnv_AWSEnv{
				Name: "test-environment",
				Spec: &sdk.AWSEnvSpecFragment{
					Cidr:                         "10.0.0.0/16",
					Region:                       "us-east-1",
					Nat:                          true,
					AWSAccountID:                 "123456789012",
					CustomDomain:                 &[]string{"custom.example.com"}[0],
					LoadBalancingStrategy:        sdk.LoadBalancingStrategyRoundRobin,
					ResourcePrefix:               "altinity-test-",
					PermissionsBoundaryPolicyArn: &[]string{"arn:aws:iam::123456789012:policy/boundary"}[0],
					Zones:                        []string{"us-east-1a", "us-east-1b", "us-east-1c"},
					LoadBalancers: sdk.AWSEnvSpecFragment_LoadBalancers{
						Public: sdk.AWSEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
							CrossZone:      false,
						},
						Internal: sdk.AWSEnvSpecFragment_LoadBalancers_Internal{
							Enabled:                          true,
							SourceIPRanges:                   []string{"10.0.0.0/8"},
							CrossZone:                        true,
							EndpointServiceAllowedPrincipals: []string{"arn:aws:iam::123456789012:root"},
						},
					},
					NodeGroups: []*sdk.AWSEnvSpecFragment_NodeGroups{
						{
							Name:            "system-group",
							NodeType:        "system",
							CapacityPerZone: 3,
							Zones:           []string{"us-east-1a", "us-east-1b"},
							Reservations:    []sdk.NodeReservation{},
						},
						{
							Name:            "user-group",
							NodeType:        "user",
							CapacityPerZone: 5,
							Zones:           []string{"us-east-1a"},
							Reservations:    []sdk.NodeReservation{},
						},
					},
					MaintenanceWindows: []*sdk.AWSEnvSpecFragment_MaintenanceWindows{
						{
							Name:          "weekly-maintenance",
							Enabled:       true,
							Hour:          2,
							LengthInHours: 4,
							Days:          []sdk.Day{"saturday", "sunday"},
						},
					},
					PeeringConnections: []*sdk.AWSEnvSpecFragment_PeeringConnections{
						{
							AWSAccountID: &[]string{"987654321098"}[0],
							VpcID:        "vpc-12345",
							VpcRegion:    &[]string{"us-west-2"}[0],
						},
					},
					Endpoints: []*sdk.AWSEnvSpecFragment_Endpoints{
						{
							ServiceName: "com.amazonaws.vpce.us-east-1.s3",
							Alias:       &[]string{"s3-endpoint"}[0],
							PrivateDNS:  true,
						},
					},
					Tags: []*sdk.AWSEnvSpecFragment_Tags{
						{
							Key:   "Environment",
							Value: "production",
						},
						{
							Key:   "Team",
							Value: "backend",
						},
					},
					CloudConnect: true,
				},
				SpecRevision: 42,
			},
			validate: func(t *testing.T, model *AWSEnvResourceModel) {
				if model.Name.ValueString() != "test-environment" {
					t.Errorf("Name: expected 'test-environment', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "10.0.0.0/16" {
					t.Errorf("CIDR: expected '10.0.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.Region.ValueString() != "us-east-1" {
					t.Errorf("Region: expected 'us-east-1', got '%s'", model.Region.ValueString())
				}
				if !model.NAT.ValueBool() {
					t.Errorf("NAT: expected true, got %v", model.NAT.ValueBool())
				}
				if model.AWSAccountID.ValueString() != "123456789012" {
					t.Errorf("AWSAccountID: expected '123456789012', got '%s'", model.AWSAccountID.ValueString())
				}
				if model.CustomDomain.ValueString() != "custom.example.com" {
					t.Errorf("CustomDomain: expected 'custom.example.com', got '%s'", model.CustomDomain.ValueString())
				}
				if model.LoadBalancingStrategy.ValueString() != "ROUND_ROBIN" {
					t.Errorf("LoadBalancingStrategy: expected 'ROUND_ROBIN', got '%s'", model.LoadBalancingStrategy.ValueString())
				}
				if model.ResourcePrefix.ValueString() != "altinity-test-" {
					t.Errorf("ResourcePrefix: expected 'altinity-test-', got '%s'", model.ResourcePrefix.ValueString())
				}
				if model.PermissionsBoundaryPolicyArn.ValueString() != "arn:aws:iam::123456789012:policy/boundary" {
					t.Errorf("PermissionsBoundaryPolicyArn: expected 'arn:aws:iam::123456789012:policy/boundary', got '%s'", model.PermissionsBoundaryPolicyArn.ValueString())
				}
				if model.CloudConnect.ValueBool() != true {
					t.Errorf("CloudConnect: expected true, got %v", model.CloudConnect.ValueBool())
				}
				if model.SpecRevision.ValueInt64() != 42 {
					t.Errorf("SpecRevision: expected 42, got %d", model.SpecRevision.ValueInt64())
				}

				var zones []string
				model.Zones.ElementsAs(context.TODO(), &zones, false)
				if len(zones) != 3 {
					t.Errorf("Zones count: expected 3, got %d", len(zones))
				}

				if len(model.NodeGroups) != 2 {
					t.Errorf("NodeGroups count: expected 2, got %d", len(model.NodeGroups))
				}
				if model.NodeGroups[0].Name.ValueString() != "system-group" {
					t.Errorf("First node group name: expected 'system-group', got '%s'", model.NodeGroups[0].Name.ValueString())
				}

				if len(model.MaintenanceWindows) != 1 {
					t.Errorf("MaintenanceWindows count: expected 1, got %d", len(model.MaintenanceWindows))
				}
				if model.MaintenanceWindows[0].Name.ValueString() != "weekly-maintenance" {
					t.Errorf("Maintenance window name: expected 'weekly-maintenance', got '%s'", model.MaintenanceWindows[0].Name.ValueString())
				}

				if len(model.PeeringConnections) != 1 {
					t.Errorf("PeeringConnections count: expected 1, got %d", len(model.PeeringConnections))
				}
				if model.PeeringConnections[0].VpcID.ValueString() != "vpc-12345" {
					t.Errorf("Peering connection VPC ID: expected 'vpc-12345', got '%s'", model.PeeringConnections[0].VpcID.ValueString())
				}

				if len(model.Endpoints) != 1 {
					t.Errorf("Endpoints count: expected 1, got %d", len(model.Endpoints))
				}
				if model.Endpoints[0].ServiceName.ValueString() != "com.amazonaws.vpce.us-east-1.s3" {
					t.Errorf("Endpoint service name: expected 'com.amazonaws.vpce.us-east-1.s3', got '%s'", model.Endpoints[0].ServiceName.ValueString())
				}

				if len(model.Tags) != 2 {
					t.Errorf("Tags count: expected 2, got %d", len(model.Tags))
				}
				if model.Tags[0].Key.ValueString() != "Environment" {
					t.Errorf("First tag key: expected 'Environment', got '%s'", model.Tags[0].Key.ValueString())
				}

				if model.LoadBalancers == nil {
					t.Fatal("LoadBalancers should not be nil")
				}
				if model.LoadBalancers.Public == nil {
					t.Fatal("Public load balancer should not be nil")
				}
				if !model.LoadBalancers.Public.Enabled.ValueBool() {
					t.Errorf("Public load balancer enabled: expected true, got %v", model.LoadBalancers.Public.Enabled.ValueBool())
				}
			},
		},
		{
			name: "Minimal SDK response with required fields only",
			input: sdk.GetAWSEnv_AWSEnv{
				Name: "minimal-env",
				Spec: &sdk.AWSEnvSpecFragment{
					Cidr:                  "172.16.0.0/16",
					Region:                "us-west-2",
					Nat:                   false,
					AWSAccountID:          "111122223333",
					LoadBalancingStrategy: sdk.LoadBalancingStrategyRoundRobin,
					ResourcePrefix:        "altinity-",
					Zones:                 []string{"us-west-2a"},
					LoadBalancers: sdk.AWSEnvSpecFragment_LoadBalancers{
						Public: sdk.AWSEnvSpecFragment_LoadBalancers_Public{
							Enabled:        false,
							SourceIPRanges: []string{},
							CrossZone:      false,
						},
						Internal: sdk.AWSEnvSpecFragment_LoadBalancers_Internal{
							Enabled:                          false,
							SourceIPRanges:                   []string{},
							CrossZone:                        false,
							EndpointServiceAllowedPrincipals: []string{},
						},
					},
					NodeGroups:         []*sdk.AWSEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*sdk.AWSEnvSpecFragment_MaintenanceWindows{},
					PeeringConnections: []*sdk.AWSEnvSpecFragment_PeeringConnections{},
					Endpoints:          []*sdk.AWSEnvSpecFragment_Endpoints{},
					Tags:               []*sdk.AWSEnvSpecFragment_Tags{},
					CloudConnect:       false,
				},
				SpecRevision: 1,
			},
			validate: func(t *testing.T, model *AWSEnvResourceModel) {
				if model.Name.ValueString() != "minimal-env" {
					t.Errorf("Name: expected 'minimal-env', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "172.16.0.0/16" {
					t.Errorf("CIDR: expected '172.16.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.Region.ValueString() != "us-west-2" {
					t.Errorf("Region: expected 'us-west-2', got '%s'", model.Region.ValueString())
				}
				if model.NAT.ValueBool() {
					t.Errorf("NAT: expected false, got %v", model.NAT.ValueBool())
				}
				if model.CloudConnect.ValueBool() {
					t.Errorf("CloudConnect: expected false, got %v", model.CloudConnect.ValueBool())
				}
				if model.SpecRevision.ValueInt64() != 1 {
					t.Errorf("SpecRevision: expected 1, got %d", model.SpecRevision.ValueInt64())
				}

				if len(model.NodeGroups) != 0 {
					t.Errorf("NodeGroups: expected empty slice, got %d items", len(model.NodeGroups))
				}
				if len(model.MaintenanceWindows) != 0 {
					t.Errorf("MaintenanceWindows: expected empty slice, got %d items", len(model.MaintenanceWindows))
				}
				if len(model.PeeringConnections) != 0 {
					t.Errorf("PeeringConnections: expected empty slice, got %d items", len(model.PeeringConnections))
				}
				if len(model.Endpoints) != 0 {
					t.Errorf("Endpoints: expected empty slice, got %d items", len(model.Endpoints))
				}
				if len(model.Tags) != 0 {
					t.Errorf("Tags: expected empty slice, got %d items", len(model.Tags))
				}
			},
		},
		{
			name: "SDK response with nil optional fields",
			input: sdk.GetAWSEnv_AWSEnv{
				Name: "nil-fields-env",
				Spec: &sdk.AWSEnvSpecFragment{
					Cidr:                         "192.168.0.0/16",
					Region:                       "eu-west-1",
					Nat:                          true,
					AWSAccountID:                 "444455556666",
					CustomDomain:                 nil,
					LoadBalancingStrategy:        sdk.LoadBalancingStrategyRoundRobin,
					ResourcePrefix:               "test-",
					PermissionsBoundaryPolicyArn: nil,
					Zones:                        []string{"eu-west-1a", "eu-west-1b"},
					LoadBalancers: sdk.AWSEnvSpecFragment_LoadBalancers{
						Public: sdk.AWSEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0"},
							CrossZone:      true,
						},
						Internal: sdk.AWSEnvSpecFragment_LoadBalancers_Internal{
							Enabled:                          false,
							SourceIPRanges:                   []string{},
							CrossZone:                        false,
							EndpointServiceAllowedPrincipals: []string{},
						},
					},
					NodeGroups:         []*sdk.AWSEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*sdk.AWSEnvSpecFragment_MaintenanceWindows{},
					PeeringConnections: []*sdk.AWSEnvSpecFragment_PeeringConnections{},
					Endpoints:          []*sdk.AWSEnvSpecFragment_Endpoints{},
					Tags:               []*sdk.AWSEnvSpecFragment_Tags{},
					CloudConnect:       true,
				},
				SpecRevision: 5,
			},
			validate: func(t *testing.T, model *AWSEnvResourceModel) {
				if model.Name.ValueString() != "nil-fields-env" {
					t.Errorf("Name: expected 'nil-fields-env', got '%s'", model.Name.ValueString())
				}
				if !model.CustomDomain.IsNull() {
					t.Errorf("CustomDomain: expected null, got '%s'", model.CustomDomain.ValueString())
				}
				if !model.PermissionsBoundaryPolicyArn.IsNull() {
					t.Errorf("PermissionsBoundaryPolicyArn: expected null, got '%s'", model.PermissionsBoundaryPolicyArn.ValueString())
				}
				if model.CloudConnect.ValueBool() != true {
					t.Errorf("CloudConnect: expected true, got %v", model.CloudConnect.ValueBool())
				}
				if model.SpecRevision.ValueInt64() != 5 {
					t.Errorf("SpecRevision: expected 5, got %d", model.SpecRevision.ValueInt64())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &AWSEnvResourceModel{}
			model.toModel(tt.input)
			tt.validate(t, model)
		})
	}
}
