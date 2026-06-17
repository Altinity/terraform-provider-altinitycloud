//go:build e2e

package env_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// TestE2EAltinityCloudEnvAWS drives create -> update against the dev control
// plane (https://anywhere.dev.altinity.cloud) using a dummy-prefixed env, and
// asserts there is no drift after each apply (a non-empty plan = drift).
//
// The config exercises every settable field of the AWS env resource so the
// drift check validates the full spec round-trip (toSDK -> API -> toModel).
//
// Teardown is intentionally skipped: deleting an env on dev requires MFA
// confirmation, which CI cannot provide. Dummy-prefixed envs are cleaned up out
// of band.
func TestE2EAltinityCloudEnvAWS(t *testing.T) {
	test.E2EPreCheck(t)
	tf, workdir := test.NewE2ETerraform(t)
	ctx := context.Background()

	envName := "dummy-e2e-aws-" + test.GenerateRandomResourceName()

	// 1. Create with the full field set.
	test.WriteE2EConfig(t, workdir, awsE2EConfig(envName, 1))
	if err := tf.Apply(ctx); err != nil {
		t.Fatalf("create apply failed: %s", err)
	}

	// 2. Drift check: a plan right after create must be empty.
	if changed, err := tf.Plan(ctx); err != nil {
		t.Fatalf("plan after create failed: %s", err)
	} else if changed {
		t.Fatalf("unexpected drift after create for env %s", envName)
	}

	// 3. Update (capacity 1 -> 3), then assert no drift.
	test.WriteE2EConfig(t, workdir, awsE2EConfig(envName, 3))
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

// awsE2EConfig returns an AWS env resource that sets every settable attribute.
// capacity drives the mutable change exercised by the update step.
//
// Intentionally omitted (require real infrastructure the dev sandbox rejects):
//   - resource_prefix: server-assigned; setting it requires a sandbox-derived
//     prefix tied to the env name.
//   - permissions_boundary_policy_arn: requires resource_prefix + a real policy.
//   - datadog: enabling it requires an encrypted API key (enc_api_key).
//   - custom_domain: superseded here by custom_domains.
func awsE2EConfig(envName string, capacity int) string {
	return fmt.Sprintf(`
resource "%s" "dummy" {
  name           = "%s"
  cidr           = "10.0.0.0/16"
  region         = "us-east-1"
  aws_account_id = "123456789012"
  zones          = ["us-east-1a", "us-east-1b"]

  custom_domains          = ["e2e.example.com"]
  load_balancing_strategy = "ROUND_ROBIN"
  nat                     = true
  cloud_connect           = true
  eks_logging             = true

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
      cross_zone       = true
    }
    internal = {
      enabled                             = true
      source_ip_ranges                    = ["10.0.0.0/8"]
      cross_zone                          = true
      endpoint_service_allowed_principals = ["arn:aws:iam::123456789012:root"]
      endpoint_service_supported_regions  = ["us-east-1"]
    }
  }

  node_groups = [
    {
      zones             = ["us-east-1a"]
      node_type         = "t4g.large"
      capacity_per_zone = %d
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      zones             = ["us-east-1a", "us-east-1b"]
      node_type         = "m6i.large"
      capacity_per_zone = 2
      reservations      = ["CLICKHOUSE"]
    },
  ]

  maintenance_windows = [{
    name            = "weekly"
    enabled         = true
    hour            = 2
    length_in_hours = 4
    days            = ["MONDAY", "TUESDAY"]
  }]

  peering_connections = [{
    aws_account_id = "123456789012"
    vpc_id         = "vpc-12345678"
    vpc_region     = "us-east-1"
  }]

  endpoints = [{
    service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-12345678"
    alias        = "b-1.dummycluster.a1b2c3.c2.kafka.us-east-1.amazonaws.com"
    private_dns  = true
  }]

  tags = [{
    key   = "team"
    value = "platform"
  }]

  external_buckets = [{
    name = "my-external-bucket"
  }]

  backups = {
    custom_bucket = {
      name     = "my-backup-bucket"
      region   = "us-east-1"
      role_arn = "arn:aws:iam::123456789012:role/backup"
    }
  }

  iceberg = {
    catalogs = [{
      name                     = "catalog1"
      type                     = "S3"
      anonymous_access_enabled = false
      maintenance = {
        enabled = true
      }
      watches = [{
        table = "db.table"
      }]
    }]
  }

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
