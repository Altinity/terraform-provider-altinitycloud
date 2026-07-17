//go:build e2e

package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// Create -> update with every settable field, asserting no drift after each apply.
func TestE2EAltinityCloudEnvK8S(t *testing.T) {
	test.RunE2ELifecycle(t, "dummy-e2e-k8s-", k8sE2EConfig)
}

// capacity drives the update step; custom_domain omitted (superseded by custom_domains).
func k8sE2EConfig(envName string, capacity int) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name         = "%s"
  distribution = "CUSTOM"

  custom_domains          = ["e2e.example.com"]
  load_balancing_strategy = "ROUND_ROBIN"

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
      annotations = [{
        key   = "service.beta.kubernetes.io/aws-load-balancer-type"
        value = "nlb"
      }]
    }
    internal = {
      enabled          = true
      source_ip_ranges = ["10.0.0.0/8"]
      annotations = [{
        key   = "service.beta.kubernetes.io/aws-load-balancer-internal"
        value = "true"
      }]
    }
  }

  custom_node_types = [
    {
      name                     = "small"
      cpu_allocatable          = 4
      mem_allocatable_in_bytes = 8589934592
    },
    {
      name                     = "large"
      cpu_allocatable          = 8
      mem_allocatable_in_bytes = 17179869184
    },
  ]

  node_groups = [
    {
      node_type         = "small"
      zones             = ["us-east-1a", "us-east-1b"]
      capacity_per_zone = %d
      reservations      = ["SYSTEM", "ZOOKEEPER"]
      selector = [
        {
          key   = "workload"
          value = "system"
        },
        {
          key   = "tier"
          value = "infra"
        },
      ]
      tolerations = [
        {
          key      = "dedicated"
          value    = "system"
          operator = "EQUAL"
          effect   = "NO_SCHEDULE"
        },
        {
          key      = "priority"
          value    = "high"
          operator = "EQUAL"
          effect   = "NO_EXECUTE"
        },
      ]
    },
    {
      node_type         = "large"
      zones             = ["us-east-1b"]
      capacity_per_zone = 2
      reservations      = ["CLICKHOUSE"]
    },
  ]

  maintenance_windows = [
    {
      name            = "weekly"
      enabled         = true
      hour            = 2
      length_in_hours = 4
      days            = ["MONDAY", "TUESDAY", "WEDNESDAY"]
    },
    {
      name            = "biweekly"
      enabled         = true
      hour            = 6
      length_in_hours = 4
      days            = ["THURSDAY", "FRIDAY", "SATURDAY"]
    },
  ]

  logs = {
    storage = {
      s3 = {
        bucket_name = "my-logs-bucket"
        region      = "us-east-1"
      }
    }
  }

  metrics = {
    retention_period_in_days = 30
  }

  force_destroy                   = true
  skip_deprovision_on_destroy     = true
  allow_delete_while_disconnected = true
}
`, RESOURCE_NAME, envName, capacity)
}

// Required attrs only: covers default/null round-trips the maximal config can't.
func TestE2EAltinityCloudEnvK8SMinimal(t *testing.T) {
	test.RunE2ELifecycle(t, "dummy-e2e-k8s-min-", k8sE2EMinimalConfig)
}

func k8sE2EMinimalConfig(envName string, capacity int) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name         = "%s"
  distribution = "CUSTOM"

  node_groups = [{
    node_type         = "small"
    zones             = ["us-east-1a"]
    capacity_per_zone = %d
    reservations      = ["SYSTEM", "CLICKHOUSE", "ZOOKEEPER"]
  }]
}
`, RESOURCE_NAME, envName, capacity)
}
