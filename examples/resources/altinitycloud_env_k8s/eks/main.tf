resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "kubernetes" {
  # https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs
  config_context = "make sure provider points at the EKS you want to connect"
}

module "altinitycloud_connect" {
  source = "altinity/connect/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_k8s" "this" {
  name         = altinitycloud_env_certificate.this.env_name
  distribution = "EKS"
  // "node_groups" should match existing node groups/auto-scaling groups configuration.
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
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect" order on destroy.
    module.altinitycloud_connect
  ]
}
