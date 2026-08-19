locals {
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = local.zone_ids

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
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["CLICKHOUSE"]
    }
  ]

  // The environment IAM roles are granted access to every bucket listed here.
  // Unlike altinitycloud_env_aws, there is no permissions boundary to keep in sync.
  external_buckets = [
    { name = "acme-data-lake" },
    { name = "acme-clickhouse-s3-disk" },
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
