resource "altinitycloud_hosted_env_aws" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

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
}

data "altinitycloud_hosted_env_aws_status" "current" {
  name                           = altinitycloud_hosted_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_hosted_env_aws.this.spec_revision
}
