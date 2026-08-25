//go:build e2e

package env_test

import (
	"fmt"
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/test"
)

// TestE2EAltinityCloudEnvAWS drives create -> update against the dev control
// plane using a dummy-prefixed env, asserting no drift after each apply. The
// config exercises every settable field so the drift check validates the full
// spec round-trip (toSDK -> API -> toModel). Teardown is skipped (dev delete
// requires MFA).
func TestE2EAltinityCloudEnvAWS(t *testing.T) {
	test.RunE2ELifecycle(t, "dummy-e2e-aws-", awsE2EConfig)
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
      zones             = ["us-east-1a", "us-east-1b"]
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

  maintenance_windows = [
    {
      name            = "weekly"
      enabled         = true
      hour            = 2
      length_in_hours = 4
      days            = ["MONDAY", "TUESDAY"]
    },
    {
      name            = "weekend"
      enabled         = false
      hour            = 5
      length_in_hours = 4
      days            = ["SATURDAY", "SUNDAY"]
    },
  ]

  peering_connections = [{
    aws_account_id = "123456789012"
    vpc_id         = "vpc-12345678"
    vpc_region     = "us-east-1"
  }]

  endpoints = [
    {
      service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-12345678"
      alias        = "b-1.dummycluster.a1b2c3.c2.kafka.us-east-1.amazonaws.com"
      private_dns  = true
    },
    {
      service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-87654321"
      alias        = "b-2.dummycluster.a1b2c3.c2.kafka.us-east-1.amazonaws.com"
      private_dns  = false
    },
  ]

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

  clickhouse_keepers = [{
    name          = "keeper1"
    instance_type = "t4g.large"
    zones         = ["us-east-1a", "us-east-1b"]
    ha            = true
    stopped       = false
    disk = {
      size          = 30
      storage_class = "gp3"
    }
  }]

  clickhouse_clusters = [{
    name          = "cluster1"
    mode          = "STANDARD"
    image         = "altinity/clickhouse-server:24.8.14.10459.altinitystable"
    instance_type = "m6i.large"
    zones         = ["us-east-1a", "us-east-1b"]
    shards        = 1
    replicas      = 2
    stopped       = false

    keeper = {
      enabled = true
      name    = "keeper1"
    }

    disk = {
      size          = 100
      storage_class = "gp3"
      iops          = 3000
      throughput    = 125
    }

    additional_disks = [{
      name          = "disk1"
      size          = 50
      storage_class = "gp3"
    }]

    settings = [{
      key   = "max_concurrent_queries"
      value = "200"
    }]

    profiles = [{
      name = "readonly"
      settings = [{
        key   = "readonly"
        value = "1"
      }]
    }]

    users = [{
      name                           = "app"
      profile                        = "readonly"
      quota                          = "default"
      allowed_cidrs                  = ["10.0.0.0/8"]
      databases                      = ["analytics"]
      access_management              = false
      named_collection_control       = false
      show_named_collections         = false
      show_named_collections_secrets = false
      password_type                  = "SHA256_HEX"
      password_value                 = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"
    }]
  }]

  force_destroy                   = true
  skip_deprovision_on_destroy     = true
  allow_delete_while_disconnected = true
}
`, RESOURCE_NAME, envName, capacity) + test.E2EEnvDataSource(RESOURCE_NAME)
}
