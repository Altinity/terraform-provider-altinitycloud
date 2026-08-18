resource "altinitycloud_hosted_env_aws" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
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
        region                   = "us-east-1"
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
