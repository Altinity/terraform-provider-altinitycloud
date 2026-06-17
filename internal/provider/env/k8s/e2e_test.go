//go:build e2e

package env_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// TestE2EAltinityCloudEnvK8S drives create -> update against the dev control
// plane using a dummy-prefixed env, asserting no drift after each apply. The
// config exercises every settable field so the drift check validates the full
// spec round-trip. Teardown is skipped (dev delete requires MFA).
func TestE2EAltinityCloudEnvK8S(t *testing.T) {
	test.E2EPreCheck(t)
	tf, workdir := test.NewE2ETerraform(t)
	ctx := context.Background()

	envName := "dummy-e2e-k8s-" + test.GenerateRandomResourceName()

	test.WriteE2EConfig(t, workdir, k8sE2EConfig(envName, 1))
	if err := tf.Apply(ctx); err != nil {
		t.Fatalf("create apply failed: %s", err)
	}
	if changed, err := tf.Plan(ctx); err != nil {
		t.Fatalf("plan after create failed: %s", err)
	} else if changed {
		t.Fatalf("unexpected drift after create for env %s", envName)
	}

	test.WriteE2EConfig(t, workdir, k8sE2EConfig(envName, 3))
	if err := tf.Apply(ctx); err != nil {
		t.Fatalf("update apply failed: %s", err)
	}
	if changed, err := tf.Plan(ctx); err != nil {
		t.Fatalf("plan after update failed: %s", err)
	} else if changed {
		t.Fatalf("unexpected drift after update for env %s", envName)
	}

	t.Logf("e2e create+update+drift OK for %s (delete skipped: dev requires MFA)", envName)
}

// k8sE2EConfig returns a BYOK env resource that sets every settable attribute.
// capacity drives the mutable change exercised by the update step.
//
// Intentionally omitted: custom_domain (superseded by custom_domains).
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
      zones             = ["us-east-1a"]
      capacity_per_zone = %d
      reservations      = ["SYSTEM", "ZOOKEEPER"]
      selector = [{
        key   = "workload"
        value = "system"
      }]
      tolerations = [{
        key      = "dedicated"
        value    = "system"
        operator = "EQUAL"
        effect   = "NO_SCHEDULE"
      }]
    },
    {
      node_type         = "large"
      zones             = ["us-east-1b"]
      capacity_per_zone = 2
      reservations      = ["CLICKHOUSE"]
    },
  ]

  maintenance_windows = [{
    name            = "weekly"
    enabled         = true
    hour            = 2
    length_in_hours = 4
    days            = ["MONDAY", "TUESDAY", "WEDNESDAY"]
  }]

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
