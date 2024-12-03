resource "altinitycloud_env_hcloud" "this" {
  name             = "acme-staging"
  cidr             = "10.136.0.0/21"
  network_zone     = "us-west"
  locations        = ["hil"]
  hcloud_token_enc = "encrypted-token"

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }

  node_groups = [{
    capacity_per_location = 10
    name                  = "cpx11"
    node_type             = "cpx11"
    reservations          = ["CLICKHOUSE", "SYSTEM", "ZOOKEEPER"]
    locations             = ["hil"]
  }]
}


data "altinitycloud_env_hcloud_status" "current" {
  name                           = altinitycloud_env_gcp.this.name
  wait_for_applied_spec_revision = altinitycloud_env_gcp.this.spec_revision
}
