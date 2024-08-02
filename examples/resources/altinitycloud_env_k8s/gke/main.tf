resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "kubernetes" {
  # https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs
  config_context = "make sure provider points at the GKE you want to connect"
}

module "altinitycloud_connect" {
  source = "altinity/connect/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_k8s" "this" {
  name         = altinitycloud_env_certificate.this.env_name
  distribution = "GCP"
  // node_groups should match existing node pools configuration.
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
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect" order on destroy.
    module.altinitycloud_connect
  ]
}

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_k8s_status" "this" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
