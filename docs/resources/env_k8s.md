---
page_title: "altinitycloud_env_k8s Resource - terraform-provider-altinitycloud"
subcategory: ""
description: |-
  Bring Your Own Kubernetes (BYOK) environment resource.
---

# altinitycloud_env_k8s (Resource)

> For a detailed guide on provisioning a K8S environment using Terraform, check our official [documentation](https://docs.altinity.com/altinitycloudanywhere/bring-your-own-kubernetes-byok/terraform/).

Bring Your Own Kubernetes (BYOK) environment resource.

## Example Usage

### BYOK/EKS (AWS)
```terraform
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

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_k8s_status" "this" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
```

### BYOK/GKE (GCP):
```terraform
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
      zones             = ["us-east1-b", "us-east1-d"]
    },
    {
      node_type         = "n2d-standard-2"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
      zones             = ["us-east1-b", "us-east1-d"]
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

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_k8s_status" "this" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
```

### BYOK/AKS (Azure):
```terraform
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
      node_type         = "Standard_B2pls_v2"
      zones             = ["eastus-1", "eastus-2"]
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "Standard_B2s_v2"
      zones             = ["eastus-1", "eastus-2"]
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
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

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_k8s_status" "this" {
  name                           = altinitycloud_env_k8s.this.name
  wait_for_applied_spec_revision = altinitycloud_env_k8s.this.spec_revision
}
```

## Set up AWS EKS cluster with [Altinity's Terraform module for BYOK on EKS](https://registry.terraform.io/modules/Altinity/eks-clickhouse/aws)

The Altinity Terraform module for EKS makes it easy to set up an EKS Kubernetes cluster for a Bring Your Own Kubernetes (BYOK) environment.

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `distribution` (String) Kubernetes distribution. **[IMMUTABLE]**

		Possible values:
		- "AKS"
		- "EKS"
		- "GKE"
		- "CUSTOM"
- `name` (String) A globally-unique environment identifier. **[IMMUTABLE]**

		- All environment names must start with your account name as prefix.
		- ⚠️ Changing environment name after creation will force a resource replacement.

		Examples:
		- "acme-staging" (where "acme" is your account name)
- `node_groups` (Attributes List) List of node groups. At least one required. (see [below for nested schema](#nestedatt--node_groups))

### Optional

- `custom_domain` (String) Custom domain.

		Examples:
		- "example.com"
		- "foo.bar.com"

		Before specifying custom domain, please create the following DNS records:
		- CNAME _acme-challenge.example.com. $env_name.altinity.cloud.
		- (optional, public load balancer)
			CNAME *.example.com. _.$env_name.altinity.cloud.
		- (optional, internal load balancer)
			CNAME *.internal.example.com. _.internal.$env_name.altinity.cloud.
		- (optional, vpce)
			CNAME *.vpce.example.com. _.vpce.$env_name.altinity.cloud.
- `custom_node_types` (Attributes List) Custom node types (see [below for nested schema](#nestedatt--custom_node_types))
- `force_destroy` (Boolean) Locks the environment for accidental deletion when running `terraform destroy` command. Your environment will be deleted, only when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run (before running the `terraform destroy`) to update this value in the state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. (default `false`)
- `force_destroy_clusters` (Boolean) By default, the destroy operation will not delete any provisioned clusters and the deletion will fail until the clusters get removed. Set to `true` to remove all provisioned clusters as part of the environment deletion process.
- `load_balancers` (Attributes) Load balancers configuration. (see [below for nested schema](#nestedatt--load_balancers))
- `load_balancing_strategy` (String) Load balancing strategy for the environment.

		Possible Values:
		- "ROUND_ROBIN": load balance traffic across all zones in round-robin fashion (default)
		- "ZONE_BEST_EFFORT": keep traffic within same zone
- `logs` (Attributes) Kubernetes environment logs configuration (see [below for nested schema](#nestedatt--logs))
- `maintenance_windows` (Attributes List) List of maintenance windows during which automatic maintenance is permitted. By default updates are applied as soon as they are available. (see [below for nested schema](#nestedatt--maintenance_windows))
- `metrics` (Attributes) Metrics configuration (see [below for nested schema](#nestedatt--metrics))
- `skip_deprovision_on_destroy` (Boolean) Set to `true` will delete without waiting for environment deprovisioning. Use this with precaution, it may end up with dangling resources in your cloud provider (default `false`).

### Read-Only

- `id` (String) ID of the environment (automatically generated based on the name)
- `spec_revision` (Number) Spec revision

<a id="nestedatt--node_groups"></a>
### Nested Schema for `node_groups`

Required:

- `capacity_per_zone` (Number) Maximum number of instances per availability zone.
- `node_type` (String) node.kubernetes.io/instance-type value.
- `zones` (List of String) topology.kubernetes.io/zone values.

Optional:

- `name` (String) Unique (among environment node groups) node group identifier.
- `reservations` (Set of String) Types of workload that are allowed to be scheduled onto the nodes that belong to this group.

		Possible values:
		- "SYSTEM" (at least one node group must include a SYSTEM reservation)
		- "CLICKHOUSE"
		- "ZOOKEEPER"
- `selector` (Attributes List) `nodeSelector` to apply to the pods targeting this group (see [below for nested schema](#nestedatt--node_groups--selector))
- `tolerations` (Attributes List) List of tolerations to apply to the pods targeting this group (see [below for nested schema](#nestedatt--node_groups--tolerations))

<a id="nestedatt--node_groups--selector"></a>
### Nested Schema for `node_groups.selector`

Required:

- `key` (String) Name of the key
- `value` (String) Value of the key


<a id="nestedatt--node_groups--tolerations"></a>
### Nested Schema for `node_groups.tolerations`

Required:

- `effect` (String) Node taint effect.

		Possible values:
		- "NO_SCHEDULE"
		- "PREFER_NO_SCHEDULE"
		- "NO_EXECUTE"
- `key` (String) Taint key, e.g. 'dedicated'
- `operator` (String) Node toleration operator used to match taints.

		Possible values:
		- "EQUALS"
		- "EXISTS"
- `value` (String) Taint value, e.g. 'clickhouse'



<a id="nestedatt--custom_node_types"></a>
### Nested Schema for `custom_node_types`

Required:

- `name` (String) Custom node type unique identifier

Optional:

- `cpu_allocatable` (Number) Number of allocatable virtual cores
- `mem_allocatable_in_bytes` (Number) Amount of allocatable memory in bytes


<a id="nestedatt--load_balancers"></a>
### Nested Schema for `load_balancers`

Optional:

- `internal` (Attributes) Internal load balancer configuration. Accessible via `*.internal.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--internal))
- `public` (Attributes) Public load balancer configuration. Accessible via `*.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--public))

<a id="nestedatt--load_balancers--internal"></a>
### Nested Schema for `load_balancers.internal`

Optional:

- `annotations` (Attributes List) List of annotations for the load balancer (see [below for nested schema](#nestedatt--load_balancers--internal--annotations))
- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from (default `"0.0.0.0/0"`).

<a id="nestedatt--load_balancers--internal--annotations"></a>
### Nested Schema for `load_balancers.internal.annotations`

Required:

- `key` (String) Name of the key
- `value` (String) Value of the key



<a id="nestedatt--load_balancers--public"></a>
### Nested Schema for `load_balancers.public`

Optional:

- `annotations` (Attributes List) List of annotations for the load balancer (see [below for nested schema](#nestedatt--load_balancers--public--annotations))
- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from (default `"0.0.0.0/0"`).

<a id="nestedatt--load_balancers--public--annotations"></a>
### Nested Schema for `load_balancers.public.annotations`

Required:

- `key` (String) Name of the key
- `value` (String) Value of the key




<a id="nestedatt--logs"></a>
### Nested Schema for `logs`

Optional:

- `storage` (Attributes) Storage backend configuration (see [below for nested schema](#nestedatt--logs--storage))

<a id="nestedatt--logs--storage"></a>
### Nested Schema for `logs.storage`

Optional:

- `gcs` (Attributes) Google Cloud Storage configuration (see [below for nested schema](#nestedatt--logs--storage--gcs))
- `s3` (Attributes) Amazon S3 configuration (see [below for nested schema](#nestedatt--logs--storage--s3))

<a id="nestedatt--logs--storage--gcs"></a>
### Nested Schema for `logs.storage.gcs`

Required:

- `bucket_name` (String) Bucket name


<a id="nestedatt--logs--storage--s3"></a>
### Nested Schema for `logs.storage.s3`

Required:

- `bucket_name` (String) Bucket name
- `region` (String) AWS region ([docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Regions)). **[IMMUTABLE]**

		Examples:
		- "us-east-1"
		- "sa-east-1"




<a id="nestedatt--maintenance_windows"></a>
### Nested Schema for `maintenance_windows`

Required:

- `days` (List of String) Days on which maintenance can take place.

		Possible values:
		- "MONDAY"
		- "TUESDAY"
		- "WEDNESDAY"
		- "THURSDAY"
		- "FRIDAY"
		- "SATURDAY"
		- "SUNDAY"
- `hour` (Number) Hour of the day in [0, 23] range.
- `length_in_hours` (Number) Maintenance window length in hours. 4h min, 24h max.
- `name` (String) Maintenance window identifier

Optional:

- `enabled` (Boolean) Set to `true` if maintenance window is enabled, `false` otherwise. (default `false`)


<a id="nestedatt--metrics"></a>
### Nested Schema for `metrics`

Optional:

- `retention_period_in_days` (Number) Metrics retention period in days (default `30`).

## Import

Import is supported using the following syntax:

```shell
terraform import altinitycloud_env_k8s.this "replace-with-environment-name"
```
