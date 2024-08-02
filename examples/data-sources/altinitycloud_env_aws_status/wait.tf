resource "altinitycloud_env_aws" "this" {
  name           = "acme-staging"
  aws_account_id = "123456789012"
  region         = "us-east-1"
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"

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
}

data "altinitycloud_env_aws_status" "current" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
