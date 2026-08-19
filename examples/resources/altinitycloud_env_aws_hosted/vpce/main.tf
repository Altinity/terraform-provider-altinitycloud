locals {
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = local.region
  zone_ids = local.zone_ids

  load_balancers = {
    internal = {
      enabled                             = true
      source_ip_ranges                    = ["10.0.0.0/8"]
      endpoint_service_allowed_principals = ["arn:aws:iam::123456789012:root"]
      endpoint_service_supported_regions  = ["us-east-1", "us-west-2"]
    }
  }

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
