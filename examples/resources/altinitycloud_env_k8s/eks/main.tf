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
      zones             = ["us-east-1a", "us-east-1b"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
      zones             = ["us-east-1a", "us-east-1b"]
      tolerations = [
        {
          key      = "dedicated"
          value    = "clickhouse"
          effect   = "NO_SCHEDULE"
          operator = "EQUAL"
        }
      ]
    }
  ]
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect" order on destroy.
    module.altinitycloud_connect
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_k8s_status" "this" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
