resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

locals {
  zones = ["us-east-1a", "us-east-1b"]
}

provider "aws" {
  region = "us-east-1"
}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = "123456789012"
  region         = "us-east-1"
  zones          = local.zones
  cidr           = "10.67.0.0/21"
  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }
  node_groups = [
    {
      node_type         = "t4g.large"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["SYSTEM"]
    },
    {
      node_type         = "t4g.small"
      capacity_per_zone = 3
      zones             = local.zones
      reservations      = ["ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zones             = local.zones
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

  // The instance type of a cluster must match a node group with a CLICKHOUSE reservation.
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
        iops          = 3000
        throughput    = 125
      }

      additional_disks = [
        {
          name = "disk1"
          size = 1000
        }
      ]

      settings = [
        {
          key   = "max_concurrent_queries"
          value = "200"
        }
      ]

      profiles = [
        {
          name = "readonly"
          settings = [
            {
              key   = "readonly"
              value = "1"
            }
          ]
        }
      ]

      users = [
        {
          name           = "app"
          allowed_cidrs  = ["10.67.0.0/21"]
          databases      = ["analytics"]
          password_type  = "SHA256_HEX"
          password_value = sha256("change-me")
        },
        {
          name    = "reporting"
          profile = "readonly"
          // The digest can also be read from a Kubernetes secret that already
          // exists in the environment namespace.
          password_value_from_secret = {
            name = "clickhouse-users"
            key  = "reporting"
          }
        }
      ]
    }
  ]

  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
