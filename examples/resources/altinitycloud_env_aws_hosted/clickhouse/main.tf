locals {
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = local.region
  zone_ids = local.zone_ids

  node_groups = [
    {
      node_type         = "t4g.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM"]
    },
    {
      node_type         = "t4g.small"
      capacity_per_zone = 3
      zone_ids          = local.zone_ids
      reservations      = ["ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["CLICKHOUSE"]
    }
  ]

  // The instance type of a Keeper must match a node group with a ZOOKEEPER reservation.
  clickhouse_keepers = [
    {
      name          = "keeper"
      instance_type = "t4g.small"
      ha            = true
      disk = {
        size = 30
      }
    }
  ]

  // Altinity-hosted environments accept altinity/clickhouse-server images only.
  clickhouse_clusters = [
    {
      name          = "analytics"
      image         = "altinity/clickhouse-server:24.8.14.10459.altinitystable"
      instance_type = "m6i.large"
      shards        = 2
      replicas      = 2

      keeper = {
        name = "keeper"
      }

      disk = {
        size          = 500
        storage_class = "gp3"
      }

      settings = [
        {
          key   = "max_concurrent_queries"
          value = "200"
        }
      ]

      users = [
        {
          name           = "app"
          databases      = ["analytics"]
          password_type  = "SHA256_HEX"
          password_value = sha256("change-me")
        }
      ]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
