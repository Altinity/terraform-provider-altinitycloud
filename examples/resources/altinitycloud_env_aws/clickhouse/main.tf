resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

// SHA256 hex digests of each user's password, stored encrypted via the secret resource.
// A password is required to create a user; it can be omitted later to keep the current one.
variable "app_password_sha256_hex" {
  type      = string
  sensitive = true
}

variable "readonly_password_sha256_hex" {
  type      = string
  sensitive = true
}

resource "altinitycloud_env_secret" "app_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.app_password_sha256_hex
}

resource "altinitycloud_env_secret" "readonly_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.readonly_password_sha256_hex
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
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["CLICKHOUSE"]
    }
  ]

  // Keyed by name: renaming a key deletes that cluster or Keeper and creates a new, empty one.
  clickhouse_keepers = {
    keeper = {
      instance_type = "t4g.large"
      zones         = local.zones
      replicas      = 3
      disk = {
        size_in_gib = 30
      }
    }
  }

  clickhouse_clusters = {
    analytics = {
      image         = "altinity/clickhouse-server:24.8.14.10459.altinitystable" // omit to get the platform default
      instance_type = "m6i.large"
      zones         = local.zones
      shards        = 1
      replicas      = 2
      keeper        = "keeper" // must name one of clickhouse_keepers
      disk = {
        size_in_gib = 100
      }
      // Extra volumes join the default storage policy; names must start with "disk".
      extra_disks = {
        diskcold = {
          size_in_gib = 500
        }
      }
      server_settings = [
        {
          key   = "max_concurrent_queries"
          value = "200"
        }
      ]
      users = {
        app = {
          profile = "default"
          quota   = "default"
          networks = {
            ips = ["10.0.0.0/8"]
          }
          databases = ["analytics"]
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.app_password.secret_value
          }
        }
        // "grants" cannot be combined with "databases" or the access flags.
        readonly = {
          grants = ["GRANT SELECT ON analytics.*"]
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.readonly_password.secret_value
          }
        }
      }
      profiles = {
        default = {
          settings = [
            {
              key   = "max_execution_time"
              value = "30"
            }
          ]
        }
      }
      quotas = {
        default = {
          intervals = [
            {
              duration = 3600
              queries  = 1000
            }
          ]
        }
      }
    }
  }

  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
// Cluster endpoints (https://<cluster>.<env>.altinity.cloud:8443) come up shortly after it returns.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
