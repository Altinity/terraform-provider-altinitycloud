locals {
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = local.region
  zone_ids = local.zone_ids

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]

  iceberg = {
    catalogs = [
      {
        name                     = "analytics"
        type                     = "S3"
        custom_s3_bucket         = "acme-iceberg"
        custom_s3_bucket_path    = "warehouse"
        region                   = local.region
        anonymous_access_enabled = false

        maintenance = {
          enabled = true
        }

        watches = [
          {
            table                            = "events"
            paths_relative_to_table_location = ["data"]
          }
        ]
      }
    ]
  }
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
