resource "altinitycloud_env_k8s" "this" {
  name         = "acme-staging"
  distribution = "EKS"

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

data "altinitycloud_env_k8s_status" "current" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
