resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "aws" {
  region = "us-east-1"
}

locals {
  account_id = "123456789012"
}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = local.account_id
  region         = "us-east-1"
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"
  load_balancers = {
    internal = {
      enabled = true
      endpoint_service_allowed_principals = [
        "arn:aws:iam::${local.account_id}:root"
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

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}


# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint.html
# Use this value as `service_name` in the `aws_vpc_endpoint` resource to create the VPC Endpoint
output "vpc_endpoint_service_name" {
  value = data.altinitycloud_env_aws_status.this.load_balancers.internal.endpoint_service_name
}
