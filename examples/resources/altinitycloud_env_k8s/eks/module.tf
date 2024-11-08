locals {
  env_name                 = "acme-staging"
  region                   = "us-east-1"
  zones                    = ["${local.region}a", "${local.region}b", "${local.region}c"]
  clickhouse_instance_type = "m6i.large"
  system_instance_type     = "t3.large"
  altinity_labels          = { "altinity.cloud/use" = "anywhere" }
}

provider "aws" {
  # https://registry.terraform.io/providers/hashicorp/aws/latest/docs
  region = local.region
}

provider "kubernetes" {
  # https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs
  host                   = module.eks_clickhouse.eks_cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks_clickhouse.eks_cluster_ca_certificate)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      local.env_name,
      "--region",
      local.region
    ]
    command = "aws"
  }
}

module "eks_clickhouse" {
  source = "github.com/Altinity/terraform-aws-eks-clickhouse"

  install_clickhouse_operator = false
  install_clickhouse_cluster  = false

  eks_cluster_name       = local.env_name
  eks_region             = local.region
  eks_cidr               = "10.0.0.0/16"
  eks_availability_zones = local.zones

  eks_private_cidr = [
    "10.0.1.0/24",
    "10.0.2.0/24",
    "10.0.3.0/24"
  ]
  eks_public_cidr = [
    "10.0.101.0/24",
    "10.0.102.0/24",
    "10.0.103.0/24"
  ]

  eks_node_pools = [
    {
      name          = "clickhouse"
      instance_type = local.clickhouse_instance_type
      desired_size  = 0
      max_size      = 10
      min_size      = 0
      zones         = local.zones
      labels        = local.altinity_labels
      taints = [
        {
          key    = "dedicated"
          value  = "clickhouse"
          effect = "NO_SCHEDULE"
        }
      ]
    },
    {
      name          = "system"
      instance_type = local.system_instance_type
      desired_size  = 0
      max_size      = 10
      min_size      = 0
      zones         = local.zones
      labels        = local.altinity_labels
    }
  ]

  eks_tags = {
    CreatedBy = "mr-robot"
  }
}

module "altinitycloud_connect" {
  source = "altinity/connect/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem

  // "depends_on" is here to enforce "this module, then module.eks_clickhouse" order on destroy.
  depends_on = [module.eks_clickhouse]
}

resource "altinitycloud_env_certificate" "this" {
  env_name = local.env_name
}

resource "altinitycloud_env_k8s" "this" {
  name         = altinitycloud_env_certificate.this.env_name
  distribution = "EKS"

  node_groups = [
    {
      name              = local.clickhouse_instance_type,
      node_type         = local.clickhouse_instance_type,
      capacity_per_zone = 10,
      reservations      = ["CLICKHOUSE"],
      zones             = local.zones
      tolerations = [
        {
          key      = "dedicated"
          value    = "clickhouse"
          effect   = "NO_SCHEDULE"
          operator = "EQUAL"
        }
      ]
    },
    {
      name              = local.system_instance_type,
      node_type         = local.system_instance_type,
      capacity_per_zone = 10,
      reservations      = ["SYSTEM", "ZOOKEEPER"],
      zones             = local.zones
    }
  ]

  load_balancers = {
    public = {
      enabled = true
    }
  }

  // "depends_on" is here to enforce "this resource, then module.altinitycloud_connect" order on destroy.
  depends_on = [module.altinitycloud_connect]
}
