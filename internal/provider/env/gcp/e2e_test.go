//go:build e2e

package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// TestE2EAltinityCloudEnvGCP drives create -> update against the dev control
// plane using a dummy-prefixed env, asserting no drift after each apply. The
// config exercises every settable field so the drift check validates the full
// spec round-trip. Teardown is skipped (dev delete requires MFA).
func TestE2EAltinityCloudEnvGCP(t *testing.T) {
	test.RunE2ELifecycle(t, "dummy-e2e-gcp-", gcpE2EConfig)
}

// gcpE2EConfig returns a GCP env resource that sets every settable attribute.
// capacity drives the mutable change exercised by the update step.
//
// Intentionally omitted: datadog (enabling requires an encrypted API key);
// custom_domain (superseded by custom_domains).
func gcpE2EConfig(envName string, capacity int) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name           = "%s"
  cidr           = "10.0.0.0/16"
  region         = "us-east1"
  gcp_project_id = "dummy-project-e2e"
  zones          = ["us-east1-b", "us-east1-c"]

  custom_domains          = ["e2e.example.com"]
  load_balancing_strategy = "ROUND_ROBIN"

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
    internal = {
      enabled          = true
      source_ip_ranges = ["10.0.0.0/8"]
    }
  }

  node_groups = [
    {
      zones             = ["us-east1-b", "us-east1-c"]
      node_type         = "c2-standard-16"
      capacity_per_zone = %d
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      zones             = ["us-east1-c"]
      node_type         = "n2-standard-8"
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
      enabled         = false
      hour            = 6
      length_in_hours = 4
      days            = ["THURSDAY", "FRIDAY", "SATURDAY"]
    },
  ]

  peering_connections = [{
    project_id   = "dummy-project-e2e"
    network_name = "my-network"
  }]

  private_service_consumers = ["dummy-consumer-proj"]

  labels = [{
    key   = "team"
    value = "platform"
  }]

  metrics_endpoint = {
    enabled          = true
    source_ip_ranges = ["0.0.0.0/0"]
  }

  force_destroy                   = true
  skip_deprovision_on_destroy     = true
  allow_delete_while_disconnected = true
}
`, RESOURCE_NAME, envName, capacity)
}
