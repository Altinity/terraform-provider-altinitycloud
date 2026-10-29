resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

// SHA256 hex digest of the password, e.g. `echo -n 'secret' | sha256sum`.
variable "admin_password_sha256_hex" {
  type      = string
  sensitive = true
}

resource "altinitycloud_env_secret" "admin_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.admin_password_sha256_hex
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

  clickhouse_clusters = {
    hello = {
      instance_type = "m6i.large" // must match a node group with the CLICKHOUSE reservation
      disk = {
        size_in_gib = 100
      }
      users = {
        admin = {
          // Without networks.ips the user can only connect from localhost.
          networks = {
            ips = ["0.0.0.0/0"]
          }
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.admin_password.secret_value
          }
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

// Waits for the environment and cluster to be reconciled and reports provisioning errors.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}

output "clickhouse_endpoint" {
  value = "https://hello.${altinitycloud_env_aws.this.name}.altinity.cloud:8443"
}
