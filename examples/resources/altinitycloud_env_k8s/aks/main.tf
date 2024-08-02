resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "kubernetes" {
  # https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs
  config_context = "make sure provider points at the AKS you want to connect"
}

module "altinitycloud_connect" {
  source = "altinity/connect/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_k8s" "this" {
  name         = altinitycloud_env_certificate.this.env_name
  distribution = "AKS"
  // node_groups should match existing node pools configuration.
  node_groups = [
    {
      node_type         = "Standard_B2s_v2"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE", "SYSTEM", "ZOOKEEPER"]
    }
  ]
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect" order on destroy.
    module.altinitycloud_connect
  ]
}
