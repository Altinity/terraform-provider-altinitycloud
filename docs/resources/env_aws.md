---
page_title: "altinitycloud_env_aws Resource - terraform-provider-altinitycloud"
subcategory: ""
description: |-
  Bring Your Own Cloud (BYOC) AWS environment resource.
---

# altinitycloud_env_aws (Resource)

> For a detailed guide on provisioning an AWS environment using Terraform, check our official [documentation](https://docs.altinity.com/altinitycloud/quickstartguide/running-in-your-cloud-byoc/aws-remote-provisioning/#method-1-using-our-terraform-provider).

Bring Your Own Cloud (BYOC) AWS environment resource.

## Example Usage

### AWS environment with public Load Balancer:
```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "aws" {
  region = "us-east-1"
}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = "123456789012"
  region         = "us-east-1"
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"
  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }
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
  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
```

### AWS environment accessible over VPC Endpoint:
```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

provider "aws" {
  region = "us-east-1"
}

locals {
  account_id = "123456789012"
}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = local.account_id
  region         = "us-east-1"
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"
  load_balancers = {
    internal = {
      enabled = true
      endpoint_service_allowed_principals = [
        "arn:aws:iam::${local.account_id}:root"
      ]
    }
  }
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
  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}

# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint.html
resource "aws_vpc_endpoint" "this" {
  service_name        = data.altinitycloud_env_aws_status.this.load_balancers.internal.endpoint_service_name
  vpc_endpoint_type   = "Interface"
  vpc_id              = var.vpc_id
  subnet_ids          = var.subnet_ids
  security_group_ids  = var.security_group_ids
  private_dns_enabled = true
}
```

### AWS environment with VPC peering:
```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

locals {
  aws_account_id = "123456789012"
  region         = "us-east-1"
}

provider "aws" {
  region = local.region
}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = local.aws_account_id
  region         = local.region
  zones          = ["us-east-1a", "us-east-1b"]
  cidr           = "10.67.0.0/21"
  load_balancers = {
    internal = {
      enabled = true
    }
  }
  peering_connections = [
    {
      aws_account_id = local.aws_account_id # This only required if the VPC is it not in the same account as the environment.
      vpc_id         = "vpc-xyz"
    }
  ]
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
  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}

resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = data.altinitycloud_env_aws_status.this.peering_connections[0].id
  auto_accept               = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_account_id` (String) ID of the AWS account ([docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/console-account-id.html#w5aac11c17b5)) in which to provision AWS resources. **[IMMUTABLE]**
- `cidr` (String) VPC CIDR block from the private IPv4 address ranges as specified in RFC 1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16). At least /21 required. **[IMMUTABLE]**

		Examples:
		- "10.136.0.0/21"
		- "172.20.0.0/21"
- `name` (String) A globally-unique environment identifier. **[IMMUTABLE]**

		- All environment names must start with your account name as prefix.
		- ⚠️ Changing environment name after creation will force a resource replacement.

		Examples:
		- "acme-staging" (where "acme" is your account name)
- `node_groups` (Attributes Set) List of node groups. At least one required. (see [below for nested schema](#nestedatt--node_groups))
- `region` (String) AWS region ([docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Regions)). **[IMMUTABLE]**

		Examples:
		- "us-east-1"
		- "sa-east-1"

### Optional

- `allow_delete_while_disconnected` (Boolean) Set to `true` to allow deletion of the environment while it is disconnected from the cloud connect. If the the environment is not connected during the deletion process you will end up in a delete timeout (default `false`).
- `cloud_connect` (Boolean) `true` indicates that cloud resources are to be managed via altinity/cloud-connect and `false` means direct management (default `true`). **[IMMUTABLE]**
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
- `endpoints` (Attributes List) AWS environment VPC endpoint configuration (see [below for nested schema](#nestedatt--endpoints))
- `force_destroy` (Boolean) Locks the environment for accidental deletion when running `terraform destroy` command. Your environment will be deleted, only when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run (before running the `terraform destroy`) to update this value in the state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. (default `false`)
- `force_destroy_clusters` (Boolean) By default, the destroy operation will not delete any provisioned clusters and the deletion will fail until the clusters get removed. Set to `true` to remove all provisioned clusters as part of the environment deletion process.
- `load_balancers` (Attributes) Load balancers configuration. (see [below for nested schema](#nestedatt--load_balancers))
- `load_balancing_strategy` (String) Load balancing strategy for the environment.

		Possible Values:
		- "ROUND_ROBIN": load balance traffic across all zones in round-robin fashion (default)
		- "ZONE_BEST_EFFORT": keep traffic within same zone
- `maintenance_windows` (Attributes List) List of maintenance windows during which automatic maintenance is permitted. By default updates are applied as soon as they are available. (see [below for nested schema](#nestedatt--maintenance_windows))
- `nat` (Boolean) Enable AWS NAT Gateway. **[IMMUTABLE]**
- `peering_connections` (Attributes List) AWS environment VPC peering configuration. (see [below for nested schema](#nestedatt--peering_connections))
- `skip_deprovision_on_destroy` (Boolean) Set to `true` will delete without waiting for environment deprovisioning. Use this with precaution, it may end up with dangling resources in your cloud provider (default `false`).
- `tags` (Attributes List) Tags to apply to AWS resources. (see [below for nested schema](#nestedatt--tags))
- `zones` (List of String) Explicit list of AWS availability zones. At least 2 required.

		Examples:
		- ["us-east-1a", "us-east-1b"]
		- ["sa-east-1c", "sa-east-1d"]

### Read-Only

- `id` (String) ID of the environment (automatically generated based on the name)
- `spec_revision` (Number) Spec revision

<a id="nestedatt--node_groups"></a>
### Nested Schema for `node_groups`

Required:

- `capacity_per_zone` (Number) Maximum number of instances per availability zone.
- `node_type` (String) List of node groups. At least one required.
- `reservations` (Set of String) Types of workload that are allowed to be scheduled onto the nodes that belong to this group.

		Possible values:
		- "SYSTEM" (at least one node group must include a SYSTEM reservation)
		- "CLICKHOUSE"
		- "ZOOKEEPER"

Optional:

- `name` (String) Unique (among environment node groups) node group identifier.
- `zones` (List of String) Availability zones. Check possible available zones in your cloud provider documentation


<a id="nestedatt--endpoints"></a>
### Nested Schema for `endpoints`

Required:

- `alias` (String) By default, VPC endpoints get assigned $endpoint_service_id.$env_name.altinity.cloud DNS record. Alias allows to override DNS record name to `$alias.$env_name.altinity.cloud`.
- `service_name` (String) VPC endpoint service name in $endpoint_service_id.$region.vpce.amazonaws.com format.

Optional:

- `private_dns` (Boolean) `true` indicates whether to associate a private hosted zone with the specified VPC (default `false`).


<a id="nestedatt--load_balancers"></a>
### Nested Schema for `load_balancers`

Optional:

- `internal` (Attributes) Internal load balancer configuration. Accessible via `*.internal.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--internal))
- `public` (Attributes) Public load balancer configuration. Accessible via `*.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--public))

<a id="nestedatt--load_balancers--internal"></a>
### Nested Schema for `load_balancers.internal`

Optional:

- `cross_zone` (Boolean) `true` indicates that traffic should be distributed across all specified availability zones, `false` otherwise. (default `true`).
- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
- `endpoint_service_allowed_principals` (List of String) ARNs for AWS principals that are allowed to create VPC endpoints.

		Examples:
		- "arn:aws:iam::$account_id:root"
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from (default `"0.0.0.0/0"`).


<a id="nestedatt--load_balancers--public"></a>
### Nested Schema for `load_balancers.public`

Optional:

- `cross_zone` (Boolean) `true` indicates that traffic should be distributed across all specified availability zones, `false` otherwise. (default `true`).
- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from (default `"0.0.0.0/0"`).



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


<a id="nestedatt--peering_connections"></a>
### Nested Schema for `peering_connections`

Required:

- `vpc_id` (String) Target VPC ID.

Optional:

- `aws_account_id` (String) ID of the AWS account ([docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/console-account-id.html#w5aac11c17b5)) in which to provision AWS resources. **[IMMUTABLE]**
- `vpc_region` (String) Target VPC region (defaults to environment region).


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String) Name of the key
- `value` (String) Value of the key
## Import

Import is supported using the following syntax:

```shell
terraform import altinitycloud_env_aws.this "replace-with-environment-name"
```
