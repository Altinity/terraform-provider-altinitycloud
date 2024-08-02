resource "altinitycloud_env_gcp" "this" {
  name           = "acme-staging"
  gcp_project_id = "gcp-project-id"
  region         = "us-east1"
  zones          = ["us-east1-b", "us-east1-d"]
  cidr           = "10.67.0.0/21"

  node_groups = [
    {
      node_type         = "e2-standard-2"
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "n2d-standard-2"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
    }
  ]
}


data "altinitycloud_env_gcp_status" "current" {
  name                           = altinitycloud_env_gcp.this.name
  wait_for_applied_spec_revision = altinitycloud_env_gcp.this.spec_revision
}
