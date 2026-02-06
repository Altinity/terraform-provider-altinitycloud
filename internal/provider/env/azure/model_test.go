package env

import (
	"context"
	"testing"

	common "github.com/altinity/terraform-provider-altinitycloud/internal/provider/env/common"
	"github.com/altinity/terraform-provider-altinitycloud/internal/sdk/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReorderNodeGroups(t *testing.T) {
	tests := []struct {
		name           string
		model          []common.NodeGroupsModel
		apiNodeGroups  []*client.AzureEnvSpecFragment_NodeGroups
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
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
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
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
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
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:  "Empty model with API node groups",
			model: []common.NodeGroupsModel{},
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
				{NodeType: "system", Name: "system-group", CapacityPerZone: 1},
				{NodeType: "user", Name: "user-group", CapacityPerZone: 2},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:           "Empty inputs",
			model:          []common.NodeGroupsModel{},
			apiNodeGroups:  []*client.AzureEnvSpecFragment_NodeGroups{},
			expectedOrder:  []string{},
			expectedLength: 0,
		},
		{
			name: "Multiple new API node groups",
			model: []common.NodeGroupsModel{
				{NodeType: types.StringValue("system")},
			},
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
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
			apiNodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
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

				nodeTypeToGroup := make(map[string]*client.AzureEnvSpecFragment_NodeGroups)
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
		input    []*client.AzureEnvSpecFragment_MaintenanceWindows
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
			input: []*client.AzureEnvSpecFragment_MaintenanceWindows{
				{
					Name:          "weekly-maintenance",
					Hour:          2,
					LengthInHours: 4,
					Days:          []client.Day{"saturday", "sunday"},
				},
				{
					Name:          "daily-maintenance",
					Hour:          1,
					LengthInHours: 1,
					Days:          []client.Day{"monday", "tuesday", "wednesday"},
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
			input:       []*client.AzureEnvSpecFragment_MaintenanceWindows{},
			expectEmpty: true,
		},
		{
			name: "Single maintenance window",
			input: []*client.AzureEnvSpecFragment_MaintenanceWindows{
				{
					Name:          "nightly-backup",
					Hour:          3,
					LengthInHours: 2,
					Days:          []client.Day{"friday"},
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
		expected *client.AzureEnvLoadBalancersSpecInput
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
				},
				Internal: &InternalLoadBalancerModel{
					Enabled:        types.BoolValue(true),
					SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8")},
				},
			},
			expected: &client.AzureEnvLoadBalancersSpecInput{
				Public: &client.AzureEnvLoadBalancerPublicSpecInput{
					Enabled:        &[]bool{true}[0],
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
				},
				Internal: &client.AzureEnvLoadBalancerInternalSpecInput{
					Enabled:        &[]bool{true}[0],
					SourceIPRanges: []string{"10.0.0.0/8"},
				},
			},
		},
		{
			name: "Only public load balancer",
			input: &LoadBalancersModel{
				Public: &PublicLoadBalancerModel{
					Enabled: types.BoolValue(false),
				},
			},
			expected: &client.AzureEnvLoadBalancersSpecInput{
				Public: &client.AzureEnvLoadBalancerPublicSpecInput{
					Enabled:        &[]bool{false}[0],
					SourceIPRanges: []string{},
				},
			},
		},
		{
			name: "Only internal load balancer",
			input: &LoadBalancersModel{
				Internal: &InternalLoadBalancerModel{
					Enabled: types.BoolValue(true),
				},
			},
			expected: &client.AzureEnvLoadBalancersSpecInput{
				Internal: &client.AzureEnvLoadBalancerInternalSpecInput{
					Enabled:        &[]bool{true}[0],
					SourceIPRanges: []string{},
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
				}
			}
		})
	}
}

func TestLoadBalancersToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    client.AzureEnvSpecFragment_LoadBalancers
		expected struct {
			publicEnabled         bool
			publicSourceIPCount   int
			internalEnabled       bool
			internalSourceIPCount int
		}
	}{
		{
			name: "Complete load balancers",
			input: client.AzureEnvSpecFragment_LoadBalancers{
				Public: client.AzureEnvSpecFragment_LoadBalancers_Public{
					Enabled:        true,
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
				},
				Internal: client.AzureEnvSpecFragment_LoadBalancers_Internal{
					Enabled:        true,
					SourceIPRanges: []string{"10.0.0.0/8"},
				},
			},
			expected: struct {
				publicEnabled         bool
				publicSourceIPCount   int
				internalEnabled       bool
				internalSourceIPCount int
			}{
				publicEnabled:         true,
				publicSourceIPCount:   2,
				internalEnabled:       true,
				internalSourceIPCount: 1,
			},
		},
		{
			name: "Minimal load balancers",
			input: client.AzureEnvSpecFragment_LoadBalancers{
				Public: client.AzureEnvSpecFragment_LoadBalancers_Public{
					Enabled:        false,
					SourceIPRanges: []string{},
				},
				Internal: client.AzureEnvSpecFragment_LoadBalancers_Internal{
					Enabled:        false,
					SourceIPRanges: []string{},
				},
			},
			expected: struct {
				publicEnabled         bool
				publicSourceIPCount   int
				internalEnabled       bool
				internalSourceIPCount int
			}{
				publicEnabled:         false,
				publicSourceIPCount:   0,
				internalEnabled:       false,
				internalSourceIPCount: 0,
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
			if len(result.Internal.SourceIPRanges) != tt.expected.internalSourceIPCount {
				t.Errorf("Internal SourceIPRanges count: expected %d, got %d", tt.expected.internalSourceIPCount, len(result.Internal.SourceIPRanges))
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
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eastus-1"), types.StringValue("eastus-2")}),
					Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
				{
					Name:            types.StringValue("user-group"),
					NodeType:        types.StringValue("user"),
					CapacityPerZone: types.Int64Value(5),
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eastus-1")}),
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
					Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eastus-1")}),
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
		input    []*client.AzureEnvSpecFragment_NodeGroups
		expected []struct {
			name            string
			nodeType        string
			capacityPerZone int64
			zonesCount      int
		}
	}{
		{
			name: "Multiple node groups",
			input: []*client.AzureEnvSpecFragment_NodeGroups{
				{
					Name:            "system-group",
					NodeType:        "system",
					CapacityPerZone: 3,
					Zones:           []string{"eastus-1", "eastus-2", "eastus-3"},
				},
				{
					Name:            "user-group",
					NodeType:        "user",
					CapacityPerZone: 10,
					Zones:           []string{"eastus-1"},
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
			input: []*client.AzureEnvSpecFragment_NodeGroups{},
			expected: []struct {
				name            string
				nodeType        string
				capacityPerZone int64
				zonesCount      int
			}{},
		},
		{
			name: "Single node group",
			input: []*client.AzureEnvSpecFragment_NodeGroups{
				{
					Name:            "logging",
					NodeType:        "logging",
					CapacityPerZone: 2,
					Zones:           []string{"eastus-1", "eastus-2"},
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

func TestMetricsEndpointToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricsEndpointModel
		expected *client.MetricsEndpointSpecInput
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Complete metrics endpoint config",
			input: &MetricsEndpointModel{
				Enabled:        types.BoolValue(true),
				SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8"), types.StringValue("192.168.1.0/24")},
			},
			expected: &client.MetricsEndpointSpecInput{
				Enabled:        &[]bool{true}[0],
				SourceIPRanges: []string{"10.0.0.0/8", "192.168.1.0/24"},
			},
		},
		{
			name: "Metrics endpoint disabled with empty source IP ranges",
			input: &MetricsEndpointModel{
				Enabled:        types.BoolValue(false),
				SourceIPRanges: []types.String{},
			},
			expected: &client.MetricsEndpointSpecInput{
				Enabled:        &[]bool{false}[0],
				SourceIPRanges: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metricsEndpointToSDK(tt.input)

			if (tt.expected == nil) != (result == nil) {
				t.Errorf("Expected nil: %v, got nil: %v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected != nil && result != nil {
				if *tt.expected.Enabled != *result.Enabled {
					t.Errorf("Enabled mismatch: expected %v, got %v", *tt.expected.Enabled, *result.Enabled)
				}

				if len(tt.expected.SourceIPRanges) != len(result.SourceIPRanges) {
					t.Errorf("SourceIPRanges count mismatch: expected %d, got %d", len(tt.expected.SourceIPRanges), len(result.SourceIPRanges))
				} else {
					for i, expected := range tt.expected.SourceIPRanges {
						if expected != result.SourceIPRanges[i] {
							t.Errorf("SourceIPRanges[%d] mismatch: expected '%s', got '%s'", i, expected, result.SourceIPRanges[i])
						}
					}
				}
			}
		})
	}
}

func TestMetricsEndpointToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *client.AzureEnvSpecFragment_MetricsEndpoint
		expected *MetricsEndpointModel
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Complete metrics endpoint response",
			input: &client.AzureEnvSpecFragment_MetricsEndpoint{
				Enabled:        true,
				SourceIPRanges: []string{"10.0.0.0/8", "172.16.0.0/12"},
			},
			expected: &MetricsEndpointModel{
				Enabled:        types.BoolValue(true),
				SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8"), types.StringValue("172.16.0.0/12")},
			},
		},
		{
			name: "Metrics endpoint disabled with empty source IP ranges",
			input: &client.AzureEnvSpecFragment_MetricsEndpoint{
				Enabled:        false,
				SourceIPRanges: []string{},
			},
			expected: &MetricsEndpointModel{
				Enabled:        types.BoolValue(false),
				SourceIPRanges: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metricsEndpointToModel(tt.input)

			if (tt.expected == nil) != (result == nil) {
				t.Errorf("Expected nil: %v, got nil: %v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected != nil && result != nil {
				if tt.expected.Enabled.ValueBool() != result.Enabled.ValueBool() {
					t.Errorf("Enabled mismatch: expected %v, got %v", tt.expected.Enabled.ValueBool(), result.Enabled.ValueBool())
				}

				if len(tt.expected.SourceIPRanges) != len(result.SourceIPRanges) {
					t.Errorf("SourceIPRanges count mismatch: expected %d, got %d", len(tt.expected.SourceIPRanges), len(result.SourceIPRanges))
				} else {
					for i, expected := range tt.expected.SourceIPRanges {
						if expected.ValueString() != result.SourceIPRanges[i].ValueString() {
							t.Errorf("SourceIPRanges[%d] mismatch: expected '%s', got '%s'", i, expected.ValueString(), result.SourceIPRanges[i].ValueString())
						}
					}
				}
			}
		})
	}
}

func TestAzureEnvResourceModel_toSDK(t *testing.T) {
	tests := []struct {
		name     string
		model    AzureEnvResourceModel
		validate func(t *testing.T, create client.CreateAzureEnvInput, update client.UpdateAzureEnvInput)
	}{
		{
			name: "Complete model with all fields",
			model: AzureEnvResourceModel{
				Name:                  types.StringValue("test-azure-env"),
				CustomDomain:          types.StringValue("custom.azure.example.com"),
				LoadBalancingStrategy: types.StringValue("round_robin"),
				Region:                types.StringValue("East US"),
				CIDR:                  types.StringValue("10.0.0.0/16"),
				TenantID:              types.StringValue("tenant-12345"),
				SubscriptionID:        types.StringValue("subscription-67890"),
				Zones:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eastus-1"), types.StringValue("eastus-2")}),
				LoadBalancers: &LoadBalancersModel{
					Public: &PublicLoadBalancerModel{
						Enabled:        types.BoolValue(true),
						SourceIPRanges: []types.String{types.StringValue("0.0.0.0/0")},
					},
				},
				NodeGroups: []common.NodeGroupsModel{
					{
						Name:            types.StringValue("system"),
						NodeType:        types.StringValue("system"),
						CapacityPerZone: types.Int64Value(2),
						Zones:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("eastus-1")}),
						Reservations:    types.SetValueMust(types.ObjectType{}, []attr.Value{}),
					},
				},
				MaintenanceWindows: []common.MaintenanceWindowModel{
					{
						Name:          types.StringValue("weekly"),
						Enabled:       types.BoolValue(true),
						Hour:          types.Int64Value(2),
						LengthInHours: types.Int64Value(4),
						Days:          []types.String{types.StringValue("saturday")},
					},
				},
				Tags: []common.KeyValueModel{
					{
						Key:   types.StringValue("Environment"),
						Value: types.StringValue("test"),
					},
				},
				PrivateLinkService: &PrivateLinkServiceModel{
					AllowedSubscriptions: []types.String{types.StringValue("sub-123"), types.StringValue("sub-456")},
				},
				MetricsEndpoint: &MetricsEndpointModel{
					Enabled:        types.BoolValue(true),
					SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8")},
				},
			},
			validate: func(t *testing.T, create client.CreateAzureEnvInput, update client.UpdateAzureEnvInput) {

				if create.Name != "test-azure-env" {
					t.Errorf("Create name: expected 'test-azure-env', got '%s'", create.Name)
				}
				if create.Spec == nil {
					t.Fatal("Create spec should not be nil")
				}
				if *create.Spec.CustomDomain != "custom.azure.example.com" {
					t.Errorf("Create custom domain: expected 'custom.azure.example.com', got '%s'", *create.Spec.CustomDomain)
				}
				if create.Spec.Region != "East US" {
					t.Errorf("Create region: expected 'East US', got '%s'", create.Spec.Region)
				}
				if create.Spec.TenantID != "tenant-12345" {
					t.Errorf("Create tenant ID: expected 'tenant-12345', got '%s'", create.Spec.TenantID)
				}
				if create.Spec.SubscriptionID != "subscription-67890" {
					t.Errorf("Create subscription ID: expected 'subscription-67890', got '%s'", create.Spec.SubscriptionID)
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
				if len(create.Spec.Tags) != 1 {
					t.Errorf("Create tags: expected 1, got %d", len(create.Spec.Tags))
				}
				if len(create.Spec.PrivateLinkService.AllowedSubscriptions) != 2 {
					t.Errorf("Create private link allowed subscriptions: expected 2, got %d", len(create.Spec.PrivateLinkService.AllowedSubscriptions))
				}
				if create.Spec.MetricsEndpoint == nil {
					t.Fatal("Create MetricsEndpoint should not be nil")
				}
				if *create.Spec.MetricsEndpoint.Enabled != true {
					t.Errorf("Create MetricsEndpoint enabled: expected true, got %v", *create.Spec.MetricsEndpoint.Enabled)
				}

				if update.Name != "test-azure-env" {
					t.Errorf("Update name: expected 'test-azure-env', got '%s'", update.Name)
				}
				if update.Spec == nil {
					t.Fatal("Update spec should not be nil")
				}
				if *update.Spec.CustomDomain != "custom.azure.example.com" {
					t.Errorf("Update custom domain: expected 'custom.azure.example.com', got '%s'", *update.Spec.CustomDomain)
				}
			},
		},
		{
			name: "Minimal model with required fields only",
			model: AzureEnvResourceModel{
				Name:                  types.StringValue("minimal-azure-env"),
				Region:                types.StringValue("West US"),
				CIDR:                  types.StringValue("172.16.0.0/16"),
				TenantID:              types.StringValue("tenant-min"),
				SubscriptionID:        types.StringValue("subscription-min"),
				LoadBalancingStrategy: types.StringValue("zone_best_effort"),
				Zones:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("westus-1")}),
				NodeGroups:            []common.NodeGroupsModel{},
			},
			validate: func(t *testing.T, create client.CreateAzureEnvInput, update client.UpdateAzureEnvInput) {
				if create.Name != "minimal-azure-env" {
					t.Errorf("Create name: expected 'minimal-azure-env', got '%s'", create.Name)
				}
				if create.Spec.Region != "West US" {
					t.Errorf("Create region: expected 'West US', got '%s'", create.Spec.Region)
				}
				if create.Spec.Cidr != "172.16.0.0/16" {
					t.Errorf("Create CIDR: expected '172.16.0.0/16', got '%s'", create.Spec.Cidr)
				}
				if create.Spec.TenantID != "tenant-min" {
					t.Errorf("Create tenant ID: expected 'tenant-min', got '%s'", create.Spec.TenantID)
				}
				if create.Spec.SubscriptionID != "subscription-min" {
					t.Errorf("Create subscription ID: expected 'subscription-min', got '%s'", create.Spec.SubscriptionID)
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
			model: AzureEnvResourceModel{
				Name:               types.StringValue("empty-slices"),
				Region:             types.StringValue("North Europe"),
				CIDR:               types.StringValue("192.168.0.0/16"),
				TenantID:           types.StringValue("tenant-empty"),
				SubscriptionID:     types.StringValue("subscription-empty"),
				Zones:              types.ListValueMust(types.StringType, []attr.Value{types.StringValue("northeurope-1")}),
				NodeGroups:         []common.NodeGroupsModel{},
				MaintenanceWindows: []common.MaintenanceWindowModel{},
				Tags:               []common.KeyValueModel{},
			},
			validate: func(t *testing.T, create client.CreateAzureEnvInput, update client.UpdateAzureEnvInput) {
				if len(create.Spec.NodeGroups) != 0 {
					t.Errorf("Expected empty node groups, got %d", len(create.Spec.NodeGroups))
				}
				if len(create.Spec.MaintenanceWindows) != 0 {
					t.Errorf("Expected empty maintenance windows, got %d", len(create.Spec.MaintenanceWindows))
				}
				if len(create.Spec.Tags) != 0 {
					t.Errorf("Expected empty tags, got %d", len(create.Spec.Tags))
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

func TestAzureEnvResourceModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    client.GetAzureEnv_AzureEnv
		validate func(t *testing.T, model *AzureEnvResourceModel)
	}{
		{
			name: "Complete SDK response with all fields",
			input: client.GetAzureEnv_AzureEnv{
				Name: "test-azure-environment",
				Spec: &client.AzureEnvSpecFragment{
					Cidr:                  "10.0.0.0/16",
					Region:                "East US",
					TenantID:              "tenant-12345",
					SubscriptionID:        "subscription-67890",
					CustomDomain:          &[]string{"custom.azure.example.com"}[0],
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Zones:                 []string{"eastus-1", "eastus-2", "eastus-3"},
					LoadBalancers: client.AzureEnvSpecFragment_LoadBalancers{
						Public: client.AzureEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
						},
						Internal: client.AzureEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        true,
							SourceIPRanges: []string{"10.0.0.0/8"},
						},
					},
					NodeGroups: []*client.AzureEnvSpecFragment_NodeGroups{
						{
							Name:            "system-group",
							NodeType:        "system",
							CapacityPerZone: 3,
							Zones:           []string{"eastus-1", "eastus-2"},
							Reservations:    []client.NodeReservation{},
						},
						{
							Name:            "user-group",
							NodeType:        "user",
							CapacityPerZone: 5,
							Zones:           []string{"eastus-1"},
							Reservations:    []client.NodeReservation{},
						},
					},
					MaintenanceWindows: []*client.AzureEnvSpecFragment_MaintenanceWindows{
						{
							Name:          "weekly-maintenance",
							Enabled:       true,
							Hour:          2,
							LengthInHours: 4,
							Days:          []client.Day{"saturday", "sunday"},
						},
					},
					Tags: []*client.AzureEnvSpecFragment_Tags{
						{
							Key:   "Environment",
							Value: "production",
						},
						{
							Key:   "Team",
							Value: "backend",
						},
					},
					PrivateLinkService: client.AzureEnvSpecFragment_PrivateLinkService{
						AllowedSubscriptions: []string{"sub-123", "sub-456", "sub-789"},
					},
					MetricsEndpoint: client.AzureEnvSpecFragment_MetricsEndpoint{
						Enabled:        true,
						SourceIPRanges: []string{"10.0.0.0/8", "192.168.0.0/16"},
					},
				},
			},
			validate: func(t *testing.T, model *AzureEnvResourceModel) {
				if model.Name.ValueString() != "test-azure-environment" {
					t.Errorf("Name: expected 'test-azure-environment', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "10.0.0.0/16" {
					t.Errorf("CIDR: expected '10.0.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.Region.ValueString() != "East US" {
					t.Errorf("Region: expected 'East US', got '%s'", model.Region.ValueString())
				}
				if model.TenantID.ValueString() != "tenant-12345" {
					t.Errorf("TenantID: expected 'tenant-12345', got '%s'", model.TenantID.ValueString())
				}
				if model.SubscriptionID.ValueString() != "subscription-67890" {
					t.Errorf("SubscriptionID: expected 'subscription-67890', got '%s'", model.SubscriptionID.ValueString())
				}
				if model.CustomDomain.ValueString() != "custom.azure.example.com" {
					t.Errorf("CustomDomain: expected 'custom.azure.example.com', got '%s'", model.CustomDomain.ValueString())
				}
				if model.LoadBalancingStrategy.ValueString() != "ROUND_ROBIN" {
					t.Errorf("LoadBalancingStrategy: expected 'ROUND_ROBIN', got '%s'", model.LoadBalancingStrategy.ValueString())
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

				if model.PrivateLinkService == nil {
					t.Fatal("PrivateLinkService should not be nil")
				}
				if len(model.PrivateLinkService.AllowedSubscriptions) != 3 {
					t.Errorf("PrivateLinkService allowed subscriptions count: expected 3, got %d", len(model.PrivateLinkService.AllowedSubscriptions))
				}

				if model.MetricsEndpoint == nil {
					t.Fatal("MetricsEndpoint should not be nil")
				}
				if !model.MetricsEndpoint.Enabled.ValueBool() {
					t.Errorf("MetricsEndpoint enabled: expected true, got %v", model.MetricsEndpoint.Enabled.ValueBool())
				}
				if len(model.MetricsEndpoint.SourceIPRanges) != 2 {
					t.Errorf("MetricsEndpoint source IP ranges: expected 2, got %d", len(model.MetricsEndpoint.SourceIPRanges))
				}
			},
		},
		{
			name: "Minimal SDK response with required fields only",
			input: client.GetAzureEnv_AzureEnv{
				Name: "minimal-azure-env",
				Spec: &client.AzureEnvSpecFragment{
					Cidr:                  "172.16.0.0/16",
					Region:                "West US",
					TenantID:              "tenant-min",
					SubscriptionID:        "subscription-min",
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Zones:                 []string{"westus-1"},
					LoadBalancers: client.AzureEnvSpecFragment_LoadBalancers{
						Public: client.AzureEnvSpecFragment_LoadBalancers_Public{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
						Internal: client.AzureEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
					},
					NodeGroups:         []*client.AzureEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*client.AzureEnvSpecFragment_MaintenanceWindows{},
					Tags:               []*client.AzureEnvSpecFragment_Tags{},
					PrivateLinkService: client.AzureEnvSpecFragment_PrivateLinkService{
						AllowedSubscriptions: []string{},
					},
				},
			},
			validate: func(t *testing.T, model *AzureEnvResourceModel) {
				if model.Name.ValueString() != "minimal-azure-env" {
					t.Errorf("Name: expected 'minimal-azure-env', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "172.16.0.0/16" {
					t.Errorf("CIDR: expected '172.16.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.Region.ValueString() != "West US" {
					t.Errorf("Region: expected 'West US', got '%s'", model.Region.ValueString())
				}
				if model.TenantID.ValueString() != "tenant-min" {
					t.Errorf("TenantID: expected 'tenant-min', got '%s'", model.TenantID.ValueString())
				}
				if model.SubscriptionID.ValueString() != "subscription-min" {
					t.Errorf("SubscriptionID: expected 'subscription-min', got '%s'", model.SubscriptionID.ValueString())
				}

				if len(model.NodeGroups) != 0 {
					t.Errorf("NodeGroups: expected empty slice, got %d items", len(model.NodeGroups))
				}
				if len(model.MaintenanceWindows) != 0 {
					t.Errorf("MaintenanceWindows: expected empty slice, got %d items", len(model.MaintenanceWindows))
				}
				if len(model.Tags) != 0 {
					t.Errorf("Tags: expected empty slice, got %d items", len(model.Tags))
				}
				if len(model.PrivateLinkService.AllowedSubscriptions) != 0 {
					t.Errorf("PrivateLinkService allowed subscriptions: expected empty slice, got %d items", len(model.PrivateLinkService.AllowedSubscriptions))
				}
			},
		},
		{
			name: "SDK response with nil optional fields",
			input: client.GetAzureEnv_AzureEnv{
				Name: "nil-fields-azure-env",
				Spec: &client.AzureEnvSpecFragment{
					Cidr:                  "192.168.0.0/16",
					Region:                "North Europe",
					TenantID:              "tenant-nil",
					SubscriptionID:        "subscription-nil",
					CustomDomain:          nil,
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Zones:                 []string{"northeurope-1", "northeurope-2"},
					LoadBalancers: client.AzureEnvSpecFragment_LoadBalancers{
						Public: client.AzureEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0"},
						},
						Internal: client.AzureEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
					},
					NodeGroups:         []*client.AzureEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*client.AzureEnvSpecFragment_MaintenanceWindows{},
					Tags:               []*client.AzureEnvSpecFragment_Tags{},
					PrivateLinkService: client.AzureEnvSpecFragment_PrivateLinkService{
						AllowedSubscriptions: []string{},
					},
				},
			},
			validate: func(t *testing.T, model *AzureEnvResourceModel) {
				if model.Name.ValueString() != "nil-fields-azure-env" {
					t.Errorf("Name: expected 'nil-fields-azure-env', got '%s'", model.Name.ValueString())
				}
				if !model.CustomDomain.IsNull() {
					t.Errorf("CustomDomain: expected null, got '%s'", model.CustomDomain.ValueString())
				}
				if model.LoadBalancingStrategy.ValueString() != "ROUND_ROBIN" {
					t.Errorf("LoadBalancingStrategy: expected 'ROUND_ROBIN', got '%s'", model.LoadBalancingStrategy.ValueString())
				}

				var zones []string
				model.Zones.ElementsAs(context.TODO(), &zones, false)
				if len(zones) != 2 {
					t.Errorf("Zones count: expected 2, got %d", len(zones))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &AzureEnvResourceModel{}
			model.toModel(tt.input)
			tt.validate(t, model)
		})
	}
}
