resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
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
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"
  load_balancers = {
    internal = {
      enabled = true
      peering_connections = [
        {
          vcp_id = "vpc-xyz"
        }
      ]
    }
  }
  node_groups = [
    {
      node_type         = "t4g.large"
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
    }
  ]
  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}

resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = data.altinitycloud_env_aws_status.this.peering_connections[0].id
  auto_accept               = true
}
