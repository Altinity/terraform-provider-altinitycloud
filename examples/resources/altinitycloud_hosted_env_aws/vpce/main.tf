resource "altinitycloud_hosted_env_aws" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

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
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]
}
