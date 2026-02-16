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
		model          []NodeGroupsModel
		apiNodeGroups  []*client.HCloudEnvSpecFragment_NodeGroups
		expectedOrder  []string
		expectedLength int
		validateData   bool
	}{
		{
			name: "Preserve model order and add new API node groups",
			model: []NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
			},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerLocation: 2},
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 1},
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerLocation: 1},
			},
			expectedOrder:  []string{"system", "user", "monitoring"},
			expectedLength: 3,
		},
		{
			name: "All API node groups exist in model",
			model: []NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
			},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerLocation: 2},
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 1},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name: "Model has more node groups than API",
			model: []NodeGroupsModel{
				{NodeType: types.StringValue("system")},
				{NodeType: types.StringValue("user")},
				{NodeType: types.StringValue("missing")},
			},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerLocation: 2},
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 1},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:  "Empty model with API node groups",
			model: []NodeGroupsModel{},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 1},
				{NodeType: "user", Name: "user-group", CapacityPerLocation: 2},
			},
			expectedOrder:  []string{"system", "user"},
			expectedLength: 2,
		},
		{
			name:           "Empty inputs",
			model:          []NodeGroupsModel{},
			apiNodeGroups:  []*client.HCloudEnvSpecFragment_NodeGroups{},
			expectedOrder:  []string{},
			expectedLength: 0,
		},
		{
			name: "Multiple new API node groups",
			model: []NodeGroupsModel{
				{NodeType: types.StringValue("system")},
			},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerLocation: 1},
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 1},
				{NodeType: "logging", Name: "logging-group", CapacityPerLocation: 1},
				{NodeType: "metrics", Name: "metrics-group", CapacityPerLocation: 1},
			},
			expectedOrder:  []string{"system", "monitoring", "logging", "metrics"},
			expectedLength: 4,
		},
		{
			name: "No data loss validation",
			model: []NodeGroupsModel{
				{NodeType: types.StringValue("system")},
			},
			apiNodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
				{NodeType: "user", Name: "user-group", CapacityPerLocation: 5},
				{NodeType: "system", Name: "system-group", CapacityPerLocation: 10},
				{NodeType: "monitoring", Name: "monitoring-group", CapacityPerLocation: 3},
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
				if result[0].CapacityPerLocation != 10 {
					t.Errorf("Expected system CapacityPerLocation to be 10, got %d", result[0].CapacityPerLocation)
				}

				nodeTypeToGroup := make(map[string]*client.HCloudEnvSpecFragment_NodeGroups)
				for _, group := range result {
					nodeTypeToGroup[group.NodeType] = group
				}

				if nodeTypeToGroup["user"].Name != "user-group" {
					t.Errorf("Expected user Name to be 'user-group', got '%s'", nodeTypeToGroup["user"].Name)
				}
				if nodeTypeToGroup["user"].CapacityPerLocation != 5 {
					t.Errorf("Expected user CapacityPerLocation to be 5, got %d", nodeTypeToGroup["user"].CapacityPerLocation)
				}

				if nodeTypeToGroup["monitoring"].Name != "monitoring-group" {
					t.Errorf("Expected monitoring Name to be 'monitoring-group', got '%s'", nodeTypeToGroup["monitoring"].Name)
				}
				if nodeTypeToGroup["monitoring"].CapacityPerLocation != 3 {
					t.Errorf("Expected monitoring CapacityPerLocation to be 3, got %d", nodeTypeToGroup["monitoring"].CapacityPerLocation)
				}
			}
		})
	}
}

func TestMaintenanceWindowsToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    []*client.HCloudEnvSpecFragment_MaintenanceWindows
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
			input: []*client.HCloudEnvSpecFragment_MaintenanceWindows{
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
			input:       []*client.HCloudEnvSpecFragment_MaintenanceWindows{},
			expectEmpty: true,
		},
		{
			name: "Single maintenance window",
			input: []*client.HCloudEnvSpecFragment_MaintenanceWindows{
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
		expected *client.HCloudEnvLoadBalancersSpecInput
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
			expected: &client.HCloudEnvLoadBalancersSpecInput{
				Public: &client.HCloudEnvLoadBalancerPublicSpecInput{
					Enabled:        &[]bool{true}[0],
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
				},
				Internal: &client.HCloudEnvLoadBalancerInternalSpecInput{
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
			expected: &client.HCloudEnvLoadBalancersSpecInput{
				Public: &client.HCloudEnvLoadBalancerPublicSpecInput{
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
			expected: &client.HCloudEnvLoadBalancersSpecInput{
				Internal: &client.HCloudEnvLoadBalancerInternalSpecInput{
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
		input    client.HCloudEnvSpecFragment_LoadBalancers
		expected struct {
			publicEnabled         bool
			publicSourceIPCount   int
			internalEnabled       bool
			internalSourceIPCount int
		}
	}{
		{
			name: "Complete load balancers",
			input: client.HCloudEnvSpecFragment_LoadBalancers{
				Public: client.HCloudEnvSpecFragment_LoadBalancers_Public{
					Enabled:        true,
					SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
				},
				Internal: client.HCloudEnvSpecFragment_LoadBalancers_Internal{
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
			input: client.HCloudEnvSpecFragment_LoadBalancers{
				Public: client.HCloudEnvSpecFragment_LoadBalancers_Public{
					Enabled:        false,
					SourceIPRanges: []string{},
				},
				Internal: client.HCloudEnvSpecFragment_LoadBalancers_Internal{
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
		input    []NodeGroupsModel
		expected []struct {
			name                string
			nodeType            string
			capacityPerLocation int64
			locationsCount      int
		}
	}{
		{
			name: "Multiple node groups",
			input: []NodeGroupsModel{
				{
					Name:                types.StringValue("system-group"),
					NodeType:            types.StringValue("system"),
					CapacityPerLocation: types.Int64Value(2),
					Locations:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("nbg1"), types.StringValue("hel1")}),
					Reservations:        types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
				{
					Name:                types.StringValue("user-group"),
					NodeType:            types.StringValue("user"),
					CapacityPerLocation: types.Int64Value(5),
					Locations:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("fsn1")}),
					Reservations:        types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
			},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{
				{
					name:                "system-group",
					nodeType:            "system",
					capacityPerLocation: 2,
					locationsCount:      2,
				},
				{
					name:                "user-group",
					nodeType:            "user",
					capacityPerLocation: 5,
					locationsCount:      1,
				},
			},
		},
		{
			name:  "Empty input",
			input: []NodeGroupsModel{},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{},
		},
		{
			name: "Single node group",
			input: []NodeGroupsModel{
				{
					Name:                types.StringValue("monitoring"),
					NodeType:            types.StringValue("monitoring"),
					CapacityPerLocation: types.Int64Value(1),
					Locations:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ash")}),
					Reservations:        types.SetValueMust(types.ObjectType{}, []attr.Value{}),
				},
			},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{
				{
					name:                "monitoring",
					nodeType:            "monitoring",
					capacityPerLocation: 1,
					locationsCount:      1,
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
				if result[i].CapacityPerLocation != expected.capacityPerLocation {
					t.Errorf("Node group %d CapacityPerLocation: expected %d, got %d", i, expected.capacityPerLocation, result[i].CapacityPerLocation)
				}
				if len(result[i].Locations) != expected.locationsCount {
					t.Errorf("Node group %d Locations count: expected %d, got %d", i, expected.locationsCount, len(result[i].Locations))
				}
			}
		})
	}
}

func TestNodeGroupsToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    []*client.HCloudEnvSpecFragment_NodeGroups
		expected []struct {
			name                string
			nodeType            string
			capacityPerLocation int64
			locationsCount      int
		}
	}{
		{
			name: "Multiple node groups",
			input: []*client.HCloudEnvSpecFragment_NodeGroups{
				{
					Name:                "system-group",
					NodeType:            "system",
					CapacityPerLocation: 3,
					Locations:           []string{"nbg1", "hel1", "fsn1"},
				},
				{
					Name:                "user-group",
					NodeType:            "user",
					CapacityPerLocation: 10,
					Locations:           []string{"nbg1"},
				},
			},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{
				{
					name:                "system-group",
					nodeType:            "system",
					capacityPerLocation: 3,
					locationsCount:      3,
				},
				{
					name:                "user-group",
					nodeType:            "user",
					capacityPerLocation: 10,
					locationsCount:      1,
				},
			},
		},
		{
			name:  "Empty input",
			input: []*client.HCloudEnvSpecFragment_NodeGroups{},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{},
		},
		{
			name: "Single node group",
			input: []*client.HCloudEnvSpecFragment_NodeGroups{
				{
					Name:                "logging",
					NodeType:            "logging",
					CapacityPerLocation: 2,
					Locations:           []string{"hel1", "ash"},
				},
			},
			expected: []struct {
				name                string
				nodeType            string
				capacityPerLocation int64
				locationsCount      int
			}{
				{
					name:                "logging",
					nodeType:            "logging",
					capacityPerLocation: 2,
					locationsCount:      2,
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
				if result[i].CapacityPerLocation.ValueInt64() != expected.capacityPerLocation {
					t.Errorf("Node group %d CapacityPerLocation: expected %d, got %d", i, expected.capacityPerLocation, result[i].CapacityPerLocation.ValueInt64())
				}

				var locations []string
				result[i].Locations.ElementsAs(context.TODO(), &locations, false)
				if len(locations) != expected.locationsCount {
					t.Errorf("Node group %d Locations count: expected %d, got %d", i, expected.locationsCount, len(locations))
				}
			}
		})
	}
}

func TestWireguardPeersToSDK(t *testing.T) {
	tests := []struct {
		name     string
		input    []WireguardPeers
		expected []struct {
			publicKey      string
			allowedIPCount int
			endpoint       string
		}
	}{
		{
			name: "Multiple wireguard peers",
			input: []WireguardPeers{
				{
					publicKey:  types.StringValue("publickey1=="),
					allowedIPs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10.0.1.0/24"), types.StringValue("10.0.2.0/24")}),
					endpoint:   types.StringValue("1.2.3.4:51820"),
				},
				{
					publicKey:  types.StringValue("publickey2=="),
					allowedIPs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10.0.3.0/24")}),
					endpoint:   types.StringValue("5.6.7.8:51820"),
				},
			},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{
				{
					publicKey:      "publickey1==",
					allowedIPCount: 2,
					endpoint:       "1.2.3.4:51820",
				},
				{
					publicKey:      "publickey2==",
					allowedIPCount: 1,
					endpoint:       "5.6.7.8:51820",
				},
			},
		},
		{
			name:  "Empty input",
			input: []WireguardPeers{},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{},
		},
		{
			name: "Single wireguard peer",
			input: []WireguardPeers{
				{
					publicKey:  types.StringValue("singlekey=="),
					allowedIPs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.168.1.0/24")}),
					endpoint:   types.StringValue("192.168.1.1:51820"),
				},
			},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{
				{
					publicKey:      "singlekey==",
					allowedIPCount: 1,
					endpoint:       "192.168.1.1:51820",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wireguardPeersToSDK(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d wireguard peers, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].PublicKey != expected.publicKey {
					t.Errorf("Peer %d PublicKey: expected '%s', got '%s'", i, expected.publicKey, result[i].PublicKey)
				}
				if len(result[i].AllowedIPs) != expected.allowedIPCount {
					t.Errorf("Peer %d AllowedIPs count: expected %d, got %d", i, expected.allowedIPCount, len(result[i].AllowedIPs))
				}
				if result[i].Endpoint != expected.endpoint {
					t.Errorf("Peer %d Endpoint: expected '%s', got '%s'", i, expected.endpoint, result[i].Endpoint)
				}
			}
		})
	}
}

func TestWireguardPeersToModel(t *testing.T) {
	tests := []struct {
		name     string
		input    []*client.HCloudEnvSpecFragment_WireguardPeers
		expected []struct {
			publicKey      string
			allowedIPCount int
			endpoint       string
		}
	}{
		{
			name: "Multiple wireguard peers",
			input: []*client.HCloudEnvSpecFragment_WireguardPeers{
				{
					PublicKey:  "publickey1==",
					AllowedIPs: []string{"10.0.1.0/24", "10.0.2.0/24"},
					Endpoint:   "1.2.3.4:51820",
				},
				{
					PublicKey:  "publickey2==",
					AllowedIPs: []string{"10.0.3.0/24"},
					Endpoint:   "5.6.7.8:51820",
				},
			},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{
				{
					publicKey:      "publickey1==",
					allowedIPCount: 2,
					endpoint:       "1.2.3.4:51820",
				},
				{
					publicKey:      "publickey2==",
					allowedIPCount: 1,
					endpoint:       "5.6.7.8:51820",
				},
			},
		},
		{
			name:  "Empty input",
			input: []*client.HCloudEnvSpecFragment_WireguardPeers{},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{},
		},
		{
			name: "Single wireguard peer",
			input: []*client.HCloudEnvSpecFragment_WireguardPeers{
				{
					PublicKey:  "singlekey==",
					AllowedIPs: []string{"192.168.1.0/24"},
					Endpoint:   "192.168.1.1:51820",
				},
			},
			expected: []struct {
				publicKey      string
				allowedIPCount int
				endpoint       string
			}{
				{
					publicKey:      "singlekey==",
					allowedIPCount: 1,
					endpoint:       "192.168.1.1:51820",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wireguardPeersToModel(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d wireguard peers, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].publicKey.ValueString() != expected.publicKey {
					t.Errorf("Peer %d PublicKey: expected '%s', got '%s'", i, expected.publicKey, result[i].publicKey.ValueString())
				}
				if result[i].endpoint.ValueString() != expected.endpoint {
					t.Errorf("Peer %d Endpoint: expected '%s', got '%s'", i, expected.endpoint, result[i].endpoint.ValueString())
				}

				var allowedIPs []string
				result[i].allowedIPs.ElementsAs(context.TODO(), &allowedIPs, false)
				if len(allowedIPs) != expected.allowedIPCount {
					t.Errorf("Peer %d AllowedIPs count: expected %d, got %d", i, expected.allowedIPCount, len(allowedIPs))
				}
			}
		})
	}
}

func TestMetricsEndpointToSDK(t *testing.T) {
	t.Skip("metrics_endpoint temporarily removed from schema")
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
	t.Skip("metrics_endpoint temporarily removed from schema")
	tests := []struct {
		name     string
		input    *client.HCloudEnvSpecFragment_MetricsEndpoint
		expected *MetricsEndpointModel
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Complete metrics endpoint response",
			input: &client.HCloudEnvSpecFragment_MetricsEndpoint{
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
			input: &client.HCloudEnvSpecFragment_MetricsEndpoint{
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

func TestHCloudEnvResourceModel_toSDK(t *testing.T) {
	tests := []struct {
		name     string
		model    HCloudEnvResourceModel
		validate func(t *testing.T, create client.CreateHCloudEnvInput, update client.UpdateHCloudEnvInput)
	}{
		{
			name: "Complete model with all fields",
			model: HCloudEnvResourceModel{
				Name:                  types.StringValue("test-hcloud-env"),
				HCloudTokenEnc:        types.StringValue("encrypted-token-123"),
				CustomDomain:          types.StringValue("custom.hcloud.example.com"),
				LoadBalancingStrategy: types.StringValue("round_robin"),
				NetworkZone:           types.StringValue("eu-central"),
				CIDR:                  types.StringValue("10.0.0.0/16"),
				Locations:             types.ListValueMust(types.StringType, []attr.Value{types.StringValue("nbg1"), types.StringValue("hel1")}),
				LoadBalancers: &LoadBalancersModel{
					Public: &PublicLoadBalancerModel{
						Enabled:        types.BoolValue(true),
						SourceIPRanges: []types.String{types.StringValue("0.0.0.0/0")},
					},
				},
				NodeGroups: []NodeGroupsModel{
					{
						Name:                types.StringValue("system"),
						NodeType:            types.StringValue("system"),
						CapacityPerLocation: types.Int64Value(2),
						Locations:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("nbg1")}),
						Reservations:        types.SetValueMust(types.ObjectType{}, []attr.Value{}),
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
				WireguardPeers: []WireguardPeers{
					{
						publicKey:  types.StringValue("wireguard-key=="),
						allowedIPs: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("10.0.100.0/24")}),
						endpoint:   types.StringValue("1.2.3.4:51820"),
					},
				},
				// MetricsEndpoint: &MetricsEndpointModel{
				// 	Enabled:        types.BoolValue(true),
				// 	SourceIPRanges: []types.String{types.StringValue("10.0.0.0/8")},
				// },
			},
			validate: func(t *testing.T, create client.CreateHCloudEnvInput, update client.UpdateHCloudEnvInput) {
				if create.Name != "test-hcloud-env" {
					t.Errorf("Create name: expected 'test-hcloud-env', got '%s'", create.Name)
				}
				if create.Spec == nil {
					t.Fatal("Create spec should not be nil")
				}
				if create.Spec.HcloudTokenEnc != "encrypted-token-123" {
					t.Errorf("Create HCloud token: expected 'encrypted-token-123', got '%s'", create.Spec.HcloudTokenEnc)
				}
				if *create.Spec.CustomDomain != "custom.hcloud.example.com" {
					t.Errorf("Create custom domain: expected 'custom.hcloud.example.com', got '%s'", *create.Spec.CustomDomain)
				}
				if create.Spec.NetworkZone != "eu-central" {
					t.Errorf("Create network zone: expected 'eu-central', got '%s'", create.Spec.NetworkZone)
				}
				if create.Spec.Cidr != "10.0.0.0/16" {
					t.Errorf("Create CIDR: expected '10.0.0.0/16', got '%s'", create.Spec.Cidr)
				}
				if len(create.Spec.Locations) != 2 {
					t.Errorf("Create locations: expected 2, got %d", len(create.Spec.Locations))
				}
				if len(create.Spec.NodeGroups) != 1 {
					t.Errorf("Create node groups: expected 1, got %d", len(create.Spec.NodeGroups))
				}
				if len(create.Spec.WireguardPeers) != 1 {
					t.Errorf("Create wireguard peers: expected 1, got %d", len(create.Spec.WireguardPeers))
				}
				// metrics_endpoint temporarily removed from schema
				// if create.Spec.MetricsEndpoint == nil {
				// 	t.Fatal("Create MetricsEndpoint should not be nil")
				// }
				// if *create.Spec.MetricsEndpoint.Enabled != true {
				// 	t.Errorf("Create MetricsEndpoint enabled: expected true, got %v", *create.Spec.MetricsEndpoint.Enabled)
				// }

				if update.Name != "test-hcloud-env" {
					t.Errorf("Update name: expected 'test-hcloud-env', got '%s'", update.Name)
				}
				if update.Spec == nil {
					t.Fatal("Update spec should not be nil")
				}
				if *update.Spec.CustomDomain != "custom.hcloud.example.com" {
					t.Errorf("Update custom domain: expected 'custom.hcloud.example.com', got '%s'", *update.Spec.CustomDomain)
				}
			},
		},
		{
			name: "Minimal model with required fields only",
			model: HCloudEnvResourceModel{
				Name:                  types.StringValue("minimal-hcloud-env"),
				HCloudTokenEnc:        types.StringValue("minimal-token"),
				NetworkZone:           types.StringValue("eu-central"),
				CIDR:                  types.StringValue("172.16.0.0/16"),
				LoadBalancingStrategy: types.StringValue("zone_best_effort"),
				Locations:             types.ListValueMust(types.StringType, []attr.Value{types.StringValue("fsn1")}),
				NodeGroups:            []NodeGroupsModel{},
				WireguardPeers:        []WireguardPeers{},
			},
			validate: func(t *testing.T, create client.CreateHCloudEnvInput, update client.UpdateHCloudEnvInput) {
				if create.Name != "minimal-hcloud-env" {
					t.Errorf("Create name: expected 'minimal-hcloud-env', got '%s'", create.Name)
				}
				if create.Spec.NetworkZone != "eu-central" {
					t.Errorf("Create network zone: expected 'eu-central', got '%s'", create.Spec.NetworkZone)
				}
				if create.Spec.Cidr != "172.16.0.0/16" {
					t.Errorf("Create CIDR: expected '172.16.0.0/16', got '%s'", create.Spec.Cidr)
				}
				if create.Spec.HcloudTokenEnc != "minimal-token" {
					t.Errorf("Create HCloud token: expected 'minimal-token', got '%s'", create.Spec.HcloudTokenEnc)
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
			model: HCloudEnvResourceModel{
				Name:               types.StringValue("empty-slices"),
				HCloudTokenEnc:     types.StringValue("empty-token"),
				NetworkZone:        types.StringValue("us-east"),
				CIDR:               types.StringValue("192.168.0.0/16"),
				Locations:          types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ash")}),
				NodeGroups:         []NodeGroupsModel{},
				MaintenanceWindows: []common.MaintenanceWindowModel{},
				WireguardPeers:     []WireguardPeers{},
			},
			validate: func(t *testing.T, create client.CreateHCloudEnvInput, update client.UpdateHCloudEnvInput) {
				if len(create.Spec.NodeGroups) != 0 {
					t.Errorf("Expected empty node groups, got %d", len(create.Spec.NodeGroups))
				}
				if len(create.Spec.MaintenanceWindows) != 0 {
					t.Errorf("Expected empty maintenance windows, got %d", len(create.Spec.MaintenanceWindows))
				}
				if len(create.Spec.WireguardPeers) != 0 {
					t.Errorf("Expected empty wireguard peers, got %d", len(create.Spec.WireguardPeers))
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

func TestHCloudEnvResourceModel_toModel(t *testing.T) {
	tests := []struct {
		name     string
		input    client.GetHCloudEnv_HcloudEnv
		validate func(t *testing.T, model *HCloudEnvResourceModel)
	}{
		{
			name: "Complete SDK response with all fields",
			input: client.GetHCloudEnv_HcloudEnv{
				Name: "test-hcloud-environment",
				Spec: &client.HCloudEnvSpecFragment{
					Cidr:                  "10.0.0.0/16",
					NetworkZone:           "eu-central",
					CustomDomain:          &[]string{"custom.hcloud.example.com"}[0],
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Locations:             []string{"nbg1", "hel1", "fsn1"},
					LoadBalancers: client.HCloudEnvSpecFragment_LoadBalancers{
						Public: client.HCloudEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0", "192.168.1.0/24"},
						},
						Internal: client.HCloudEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        true,
							SourceIPRanges: []string{"10.0.0.0/8"},
						},
					},
					NodeGroups: []*client.HCloudEnvSpecFragment_NodeGroups{
						{
							Name:                "system-group",
							NodeType:            "system",
							CapacityPerLocation: 3,
							Locations:           []string{"nbg1", "hel1"},
							Reservations:        []client.NodeReservation{},
						},
						{
							Name:                "user-group",
							NodeType:            "user",
							CapacityPerLocation: 5,
							Locations:           []string{"nbg1"},
							Reservations:        []client.NodeReservation{},
						},
					},
					MaintenanceWindows: []*client.HCloudEnvSpecFragment_MaintenanceWindows{
						{
							Name:          "weekly-maintenance",
							Enabled:       true,
							Hour:          2,
							LengthInHours: 4,
							Days:          []client.Day{"saturday", "sunday"},
						},
					},
					WireguardPeers: []*client.HCloudEnvSpecFragment_WireguardPeers{
						{
							PublicKey:  "wireguard-key-123==",
							AllowedIPs: []string{"10.0.100.0/24", "10.0.101.0/24"},
							Endpoint:   "1.2.3.4:51820",
						},
					},
					MetricsEndpoint: client.HCloudEnvSpecFragment_MetricsEndpoint{
						Enabled:        true,
						SourceIPRanges: []string{"10.0.0.0/8", "192.168.0.0/16"},
					},
				},
			},
			validate: func(t *testing.T, model *HCloudEnvResourceModel) {
				if model.Name.ValueString() != "test-hcloud-environment" {
					t.Errorf("Name: expected 'test-hcloud-environment', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "10.0.0.0/16" {
					t.Errorf("CIDR: expected '10.0.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.NetworkZone.ValueString() != "eu-central" {
					t.Errorf("NetworkZone: expected 'eu-central', got '%s'", model.NetworkZone.ValueString())
				}
				if model.CustomDomain.ValueString() != "custom.hcloud.example.com" {
					t.Errorf("CustomDomain: expected 'custom.hcloud.example.com', got '%s'", model.CustomDomain.ValueString())
				}
				if model.LoadBalancingStrategy.ValueString() != "ROUND_ROBIN" {
					t.Errorf("LoadBalancingStrategy: expected 'ROUND_ROBIN', got '%s'", model.LoadBalancingStrategy.ValueString())
				}

				var locations []string
				model.Locations.ElementsAs(context.TODO(), &locations, false)
				if len(locations) != 3 {
					t.Errorf("Locations count: expected 3, got %d", len(locations))
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

				if len(model.WireguardPeers) != 1 {
					t.Errorf("WireguardPeers count: expected 1, got %d", len(model.WireguardPeers))
				}
				if model.WireguardPeers[0].publicKey.ValueString() != "wireguard-key-123==" {
					t.Errorf("Wireguard peer public key: expected 'wireguard-key-123==', got '%s'", model.WireguardPeers[0].publicKey.ValueString())
				}

				// metrics_endpoint temporarily removed from schema
				// if model.MetricsEndpoint == nil {
				// 	t.Fatal("MetricsEndpoint should not be nil")
				// }
				// if !model.MetricsEndpoint.Enabled.ValueBool() {
				// 	t.Errorf("MetricsEndpoint enabled: expected true, got %v", model.MetricsEndpoint.Enabled.ValueBool())
				// }
				// if len(model.MetricsEndpoint.SourceIPRanges) != 2 {
				// 	t.Errorf("MetricsEndpoint source IP ranges: expected 2, got %d", len(model.MetricsEndpoint.SourceIPRanges))
				// }

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
			input: client.GetHCloudEnv_HcloudEnv{
				Name: "minimal-hcloud-env",
				Spec: &client.HCloudEnvSpecFragment{
					Cidr:                  "172.16.0.0/16",
					NetworkZone:           "us-east",
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Locations:             []string{"ash"},
					LoadBalancers: client.HCloudEnvSpecFragment_LoadBalancers{
						Public: client.HCloudEnvSpecFragment_LoadBalancers_Public{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
						Internal: client.HCloudEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
					},
					NodeGroups:         []*client.HCloudEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*client.HCloudEnvSpecFragment_MaintenanceWindows{},
					WireguardPeers:     []*client.HCloudEnvSpecFragment_WireguardPeers{},
				},
			},
			validate: func(t *testing.T, model *HCloudEnvResourceModel) {
				if model.Name.ValueString() != "minimal-hcloud-env" {
					t.Errorf("Name: expected 'minimal-hcloud-env', got '%s'", model.Name.ValueString())
				}
				if model.CIDR.ValueString() != "172.16.0.0/16" {
					t.Errorf("CIDR: expected '172.16.0.0/16', got '%s'", model.CIDR.ValueString())
				}
				if model.NetworkZone.ValueString() != "us-east" {
					t.Errorf("NetworkZone: expected 'us-east', got '%s'", model.NetworkZone.ValueString())
				}

				if len(model.NodeGroups) != 0 {
					t.Errorf("NodeGroups: expected empty slice, got %d items", len(model.NodeGroups))
				}
				if len(model.MaintenanceWindows) != 0 {
					t.Errorf("MaintenanceWindows: expected empty slice, got %d items", len(model.MaintenanceWindows))
				}
				if len(model.WireguardPeers) != 0 {
					t.Errorf("WireguardPeers: expected empty slice, got %d items", len(model.WireguardPeers))
				}
			},
		},
		{
			name: "SDK response with nil optional fields",
			input: client.GetHCloudEnv_HcloudEnv{
				Name: "nil-fields-hcloud-env",
				Spec: &client.HCloudEnvSpecFragment{
					Cidr:                  "192.168.0.0/16",
					NetworkZone:           "eu-central",
					CustomDomain:          nil,
					LoadBalancingStrategy: client.LoadBalancingStrategyRoundRobin,
					Locations:             []string{"nbg1", "hel1"},
					LoadBalancers: client.HCloudEnvSpecFragment_LoadBalancers{
						Public: client.HCloudEnvSpecFragment_LoadBalancers_Public{
							Enabled:        true,
							SourceIPRanges: []string{"0.0.0.0/0"},
						},
						Internal: client.HCloudEnvSpecFragment_LoadBalancers_Internal{
							Enabled:        false,
							SourceIPRanges: []string{},
						},
					},
					NodeGroups:         []*client.HCloudEnvSpecFragment_NodeGroups{},
					MaintenanceWindows: []*client.HCloudEnvSpecFragment_MaintenanceWindows{},
					WireguardPeers:     []*client.HCloudEnvSpecFragment_WireguardPeers{},
				},
			},
			validate: func(t *testing.T, model *HCloudEnvResourceModel) {
				if model.Name.ValueString() != "nil-fields-hcloud-env" {
					t.Errorf("Name: expected 'nil-fields-hcloud-env', got '%s'", model.Name.ValueString())
				}
				if !model.CustomDomain.IsNull() {
					t.Errorf("CustomDomain: expected null, got '%s'", model.CustomDomain.ValueString())
				}
				if model.LoadBalancingStrategy.ValueString() != "ROUND_ROBIN" {
					t.Errorf("LoadBalancingStrategy: expected 'ROUND_ROBIN', got '%s'", model.LoadBalancingStrategy.ValueString())
				}

				var locations []string
				model.Locations.ElementsAs(context.TODO(), &locations, false)
				if len(locations) != 2 {
					t.Errorf("Locations count: expected 2, got %d", len(locations))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &HCloudEnvResourceModel{}
			model.toModel(tt.input)
			tt.validate(t, model)
		})
	}
}
