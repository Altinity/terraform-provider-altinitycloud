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
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
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
