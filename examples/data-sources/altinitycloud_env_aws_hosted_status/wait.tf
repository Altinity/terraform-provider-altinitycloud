resource "altinitycloud_env_aws_hosted" "this" {
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

data "altinitycloud_env_aws_hosted_status" "current" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
