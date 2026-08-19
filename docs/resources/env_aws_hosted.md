---
page_title: "altinitycloud_env_aws_hosted Resource - terraform-provider-altinitycloud"
subcategory: ""
description: |-
  Altinity-hosted AWS environment resource. The environment runs in an AWS account owned by Altinity.
---

# altinitycloud_env_aws_hosted (Resource)

Altinity-hosted AWS environment resource. The environment runs in an AWS account owned by Altinity.

Unlike `altinitycloud_env_aws`, an Altinity-hosted environment runs in an AWS account owned by Altinity: there is no cloud connect certificate, no AWS account id and no peering configuration. Availability zones are addressed by their [zone id](https://docs.aws.amazon.com/global-infrastructure/latest/regions/az-ids.html) (`use1-az1`), not by zone name (`us-east-1a`).

## Example Usage

### Altinity-hosted AWS environment with Public Load Balancer:
```terraform
locals {
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = local.zone_ids

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
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["CLICKHOUSE"]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
```

### Altinity-hosted AWS environment accessible over VPC Endpoint:
```terraform
resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

  load_balancers = {
    internal = {
      enabled                             = true
      source_ip_ranges                    = ["10.0.0.0/8"]
      endpoint_service_allowed_principals = ["arn:aws:iam::123456789012:root"]
      endpoint_service_supported_regions  = ["us-east-1", "us-west-2"]
    }
  }

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]
}
```

### Altinity-hosted AWS environment with access to external S3 buckets:
```terraform
locals {
  zone_ids = ["use1-az1", "use1-az2"]
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = local.zone_ids

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
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["CLICKHOUSE"]
    }
  ]

  // The environment IAM roles are granted access to every bucket listed here.
  // Unlike altinitycloud_env_aws, there is no permissions boundary to keep in sync.
  external_buckets = [
    { name = "acme-data-lake" },
    { name = "acme-clickhouse-s3-disk" },
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
```

### Altinity-hosted AWS environment with backups on a custom S3 bucket:
```terraform
locals {
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

  // Must be set explicitly: the role's trust policy has to name the environment's
  // IAM roles, which do not exist until after the environment is created.
  resource_prefix = "acme-staging"

  // Altinity's PRODUCTION AWS Organization ID. Dev/staging use a different org.
  altinity_org_id = "o-u95tkx5okz"

  backup_bucket = "acme-clickhouse-backups"
}

provider "aws" {
  region = local.region
}

resource "aws_s3_bucket" "backups" {
  bucket = local.backup_bucket
}

// The environment runs in Altinity's account, so it reaches the bucket by
// assuming this role in the customer's account rather than via its own IRSA role.
resource "aws_iam_role" "backups" {
  name = "${local.resource_prefix}-backups"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { AWS = "*" }
        Action    = "sts:AssumeRole"
        Condition = {
          StringEquals = { "aws:PrincipalOrgID" = local.altinity_org_id }
          ArnLike      = { "aws:PrincipalArn" = "arn:aws:iam::*:role/${local.resource_prefix}-*" }
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "backups" {
  role = aws_iam_role.backups.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:ListBucket", "s3:GetBucketLocation"]
        Resource = aws_s3_bucket.backups.arn
      },
      {
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:AbortMultipartUpload"]
        Resource = "${aws_s3_bucket.backups.arn}/*"
      },
    ]
  })
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = local.region
  zone_ids = local.zone_ids

  resource_prefix = local.resource_prefix

  backups = {
    custom_bucket = {
      name     = aws_s3_bucket.backups.bucket
      region   = local.region
      role_arn = aws_iam_role.backups.arn
    }
  }

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
```

### Altinity-hosted AWS environment with Iceberg catalogs:
```terraform
resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER", "CLICKHOUSE"]
    }
  ]

  iceberg = {
    catalogs = [
      {
        name                     = "analytics"
        type                     = "S3"
        custom_s3_bucket         = "acme-iceberg"
        custom_s3_bucket_path    = "warehouse"
        region                   = "us-east-1"
        anonymous_access_enabled = false

        maintenance = {
          enabled = true
        }

        watches = [
          {
            table                            = "events"
            paths_relative_to_table_location = ["data"]
          }
        ]
      }
    ]
  }
}
```

### Altinity-hosted AWS environment encrypted with a customer-managed KMS key:
```terraform
locals {
  region   = "us-east-1"
  zone_ids = ["use1-az1", "use1-az2"]

  // Must be set explicitly: the key policy has to name the environment's IAM
  // roles, which do not exist until after the environment is created.
  resource_prefix = "acme-staging"

  // Altinity's PRODUCTION AWS Organization ID. Dev/staging use a different org
  // — the env's "Encryption" page in the console renders whichever applies.
  altinity_org_id = "o-u95tkx5okz"
}

provider "aws" {
  region = local.region
}

data "aws_caller_identity" "current" {}

// Customer-managed KMS key living in the customer's account. Encrypts the EBS
// volumes and Altinity-provisioned S3 buckets of an environment that runs in
// Altinity's account, so the whole grant is cross-account.
resource "aws_kms_key" "altinity_env" {
  description             = "Altinity ${local.resource_prefix} environment encryption key"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        // Scoped by org ID (must be inside Altinity's AWS org) AND by the env's
        // role-name prefix (account wildcarded — you never need to know which
        // Altinity infra account the env lands in).
        Sid       = "AllowEnvRoles"
        Effect    = "Allow"
        Principal = { AWS = "*" }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
          "kms:CreateGrant",
          "kms:ListGrants",
          "kms:RevokeGrant",
        ]
        Resource = "*"
        Condition = {
          StringEquals = { "aws:PrincipalOrgID" = local.altinity_org_id }
          ArnLike      = { "aws:PrincipalArn" = "arn:aws:iam::*:role/${local.resource_prefix}-*" }
        }
      },
      {
        // Admin statement you author for yourself. AWS rejects a key policy
        // with no admin (lockout check), so this is required.
        Sid       = "KMSAdmins"
        Effect    = "Allow"
        Principal = { AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root" }
        Action    = "kms:*"
        Resource  = "*"
      },
    ]
  })
}

resource "altinitycloud_env_aws_hosted" "this" {
  name     = "acme-staging"
  region   = local.region
  zone_ids = local.zone_ids

  // 🚨 Both are immutable and changing them requires recreation.
  resource_prefix = local.resource_prefix
  kms_key_arn     = aws_kms_key.altinity_env.arn

  // Per-bucket key: grants the ClickHouse IRSA role decrypt/encrypt on the
  // listed external bucket so SSE-KMS objects can be read/written. Mutable.
  external_buckets = [
    {
      name        = "my-external-bucket"
      kms_key_arn = "arn:aws:kms:us-east-1:123456789012:key/66666666-7777-8888-9999-000000000000"
    }
  ]

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
      zone_ids          = local.zone_ids
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zone_ids          = local.zone_ids
      reservations      = ["CLICKHOUSE"]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_hosted_status" "this" {
  name                           = altinitycloud_env_aws_hosted.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws_hosted.this.spec_revision
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A globally-unique environment identifier. **[IMMUTABLE]**

		- All environment names must start with your account name as prefix.
		- ⚠️ Changing environment name after creation will force a resource replacement.

		Examples:
		- "acme-staging" (where "acme" is your account name)
- `node_groups` (Attributes List) List of node groups. At least one required. (see [below for nested schema](#nestedatt--node_groups))
- `region` (String) AWS region ([docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Regions)). **[IMMUTABLE]**

		Examples:
		- "us-east-1"
		- "sa-east-1"
- `zone_ids` (List of String) Explicit list of AWS availability zone ids. At least 2 required.

		Examples:
		- ["usw2-az1", "usw2-az2"]
		- ["use1-az1", "use1-az2"]

### Optional

- `allow_delete_while_disconnected` (Boolean) Set to `true` to allow deletion of the environment while it is disconnected from the cloud connect. If the the environment is not connected during the deletion process you will end up in a delete timeout (default `false`).
- `backups` (Attributes) Configuration for backup storage (see [below for nested schema](#nestedatt--backups))
- `custom_domains` (List of String) Custom domains.

		Examples:
		- "example.com"
		- "foo.bar.com"

		For each custom domain you specify, please create the following DNS records:
		`CNAME _acme-challenge.<custom_domain>. $env_name.altinity.cloud.`

		E.g. for the above examples your records should be:
		- `CNAME _acme-challenge.example.com. $env_name.altinity.cloud.`
		- `CNAME _acme-challenge.foo.bar.com. $env_name.altinity.cloud.`

		This will allow altinity to automatically provision a certificate for your custom domain.

		You should also setup a CNAME to point from your custom domain to the environment public loadbalancer:

		 - `CNAME *.<custom_domain>. _.$env_name.altinity.cloud.`

		So for the above examples you would have two additional CNAME records:

		- `CNAME *.example.com. _.$env_name.altinity.cloud.`
		- `CNAME *.foo.bar.com. _.$env_name.altinity.cloud.`
- `datadog` (Attributes) Datadog agent configuration. (see [below for nested schema](#nestedatt--datadog))
- `endpoints` (Attributes List) AWS environment VPC endpoint configuration (see [below for nested schema](#nestedatt--endpoints))
- `external_buckets` (Attributes Set) List of external S3 buckets to allow access to. The environment IAM roles are granted access to every bucket listed here. (see [below for nested schema](#nestedatt--external_buckets))
- `force_destroy` (Boolean) Locks the environment for accidental deletion when running `terraform destroy` command. Your environment will be deleted, only when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run (before running the `terraform destroy`) to update this value in the state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. (default `false`)
- `force_destroy_clusters` (Boolean) By default, the destroy operation will not delete any provisioned clusters and the deletion will fail until the clusters get removed. Set to `true` to remove all provisioned clusters as part of the environment deletion process.
- `iceberg` (Attributes) Iceberg configuration for Apache Iceberg table format support. (see [below for nested schema](#nestedatt--iceberg))
- `kms_key_arn` (String) ARN of the customer's KMS key for encrypting Altinity-provisioned data buckets and EBS volumes. **[IMMUTABLE]**
- `load_balancers` (Attributes) Load balancers configuration. (see [below for nested schema](#nestedatt--load_balancers))
- `maintenance_windows` (Attributes List) List of maintenance windows during which automatic maintenance is permitted. By default updates are applied as soon as they are available. (see [below for nested schema](#nestedatt--maintenance_windows))
- `metrics_endpoint` (Attributes) Metrics endpoint configuration. (see [below for nested schema](#nestedatt--metrics_endpoint))
- `resource_prefix` (String) Prefix applied to the names of the cloud resources created for this environment. **[IMMUTABLE]**
- `skip_deprovision_on_destroy` (Boolean) Set to `true` will delete without waiting for environment deprovisioning. Use this with precaution, it may end up with dangling resources in your cloud provider (default `false`).
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `cidr` (String) VPC CIDR block assigned to the environment.
- `id` (String) ID of the environment (automatically generated based on the name)
- `spec_revision` (Number) Spec revision

<a id="nestedatt--node_groups"></a>
### Nested Schema for `node_groups`

Required:

- `capacity_per_zone` (Number) Maximum number of instances per availability zone.
- `node_type` (String) Instance type ([docs](https://aws.amazon.com/ec2/instance-types/))

		Examples:
		- "t4g.large"
- `reservations` (Set of String) Types of workload that are allowed to be scheduled onto the nodes that belong to this group.

		Possible values:
		- "SYSTEM" (at least one node group must include a SYSTEM reservation)
		- "CLICKHOUSE"
		- "ZOOKEEPER"

Optional:

- `name` (String) Unique (among environment node groups) node group identifier.
- `zone_ids` (List of String) Availability zone ids the node group spans. Defaults to the environment zone ids.


<a id="nestedatt--backups"></a>
### Nested Schema for `backups`

Optional:

- `custom_bucket` (Attributes) Custom S3 bucket configuration for backups (see [below for nested schema](#nestedatt--backups--custom_bucket))

<a id="nestedatt--backups--custom_bucket"></a>
### Nested Schema for `backups.custom_bucket`

Required:

- `name` (String) S3 bucket name for backups
- `region` (String) AWS region where the backup bucket is located
- `role_arn` (String) Authentication configuration for backup bucket access



<a id="nestedatt--datadog"></a>
### Nested Schema for `datadog`

Optional:

- `domain` (String) Datadog intake site domain (e.g. `us3.datadoghq.com`, `app.datadoghq.eu`). Defaults to `datadoghq.com`.
- `enabled` (Boolean) Set to `true` if the Datadog agent is enabled, `false` otherwise (default `false`).
- `enc_api_key` (String, Sensitive) Datadog encrypted API key. Write-only — set to configure or rotate the key.
- `logs_enabled` (Boolean) Set to `true` to enable ClickHouse log collection, `false` otherwise (default `false`).
- `metrics_enabled` (Boolean) Set to `true` to enable ClickHouse metrics collection, `false` otherwise (default `false`).


<a id="nestedatt--endpoints"></a>
### Nested Schema for `endpoints`

Required:

- `service_name` (String) VPC endpoint service name in $endpoint_service_id.$region.vpce.amazonaws.com format.

Optional:

- `alias` (String) By default, VPC endpoints get assigned $endpoint_service_id.$env_name.altinity.cloud DNS record. Alias allows to override DNS record name to `$alias.$env_name.altinity.cloud`.


<a id="nestedatt--external_buckets"></a>
### Nested Schema for `external_buckets`

Required:

- `name` (String) External bucket name.

Optional:

- `kms_key_arn` (String) Optional ARN of a customer-managed KMS key used to encrypt this bucket. When set, the ClickHouse IRSA role is granted KMS decrypt/encrypt permissions on the key so SSE-KMS-encrypted objects in the bucket can be read and written (e.g. when the bucket backs a ClickHouse external disk). The key is owned by the customer; bucket-level encryption is not managed by Altinity. The env-region constraint that applies to the env-level KMS key does not apply here — the key may be in any region from the env's perspective. S3 still requires the key to be in the bucket's region (or to be a KMS multi-region key with a replica in the bucket's region); that is the customer's responsibility and is not validated here.


<a id="nestedatt--iceberg"></a>
### Nested Schema for `iceberg`

Required:

- `catalogs` (Attributes List) List of Iceberg catalogs. (see [below for nested schema](#nestedatt--iceberg--catalogs))

<a id="nestedatt--iceberg--catalogs"></a>
### Nested Schema for `iceberg.catalogs`

Required:

- `type` (String) Catalog type.

		Possible values:
		- "S3": S3 bucket-based catalog
		- "S3_TABLE": S3 Tables-based catalog

Optional:

- `anonymous_access_enabled` (Boolean) Whether anonymous access is enabled (default `false`).
- `custom_s3_bucket` (String) Custom S3 bucket name.
- `custom_s3_bucket_path` (String) Path within the custom S3 bucket.
- `custom_s3_table_bucket_arn` (String) ARN of the S3 Tables bucket.
- `maintenance` (Attributes) Maintenance configuration for the catalog. (see [below for nested schema](#nestedatt--iceberg--catalogs--maintenance))
- `name` (String) Catalog name. Empty name represents the default catalog.
- `region` (String) AWS region for the catalog.
- `watches` (Attributes List) Table watch configurations. (see [below for nested schema](#nestedatt--iceberg--catalogs--watches))

<a id="nestedatt--iceberg--catalogs--maintenance"></a>
### Nested Schema for `iceberg.catalogs.maintenance`

Optional:

- `enabled` (Boolean) Whether maintenance is enabled (default `true`).


<a id="nestedatt--iceberg--catalogs--watches"></a>
### Nested Schema for `iceberg.catalogs.watches`

Required:

- `table` (String) Table name to watch.

Optional:

- `paths_relative_to_table_location` (List of String) Paths relative to table location to watch.




<a id="nestedatt--load_balancers"></a>
### Nested Schema for `load_balancers`

Optional:

- `internal` (Attributes) Internal load balancer configuration. Accessible via `*.internal.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--internal))
- `public` (Attributes) Public load balancer configuration. Accessible via `*.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--public))

<a id="nestedatt--load_balancers--internal"></a>
### Nested Schema for `load_balancers.internal`

Optional:

- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
- `endpoint_service_allowed_principals` (Set of String) ARNs for AWS principals that are allowed to create VPC endpoints.

		Examples:
		- "arn:aws:iam::$account_id:root"
- `endpoint_service_supported_regions` (Set of String) List of supported regions for VPC endpoints.

		Example: ["us-east-1", "sa-east-1"]
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from (default `"0.0.0.0/0"`).


<a id="nestedatt--load_balancers--public"></a>
### Nested Schema for `load_balancers.public`

Optional:

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


<a id="nestedatt--metrics_endpoint"></a>
### Nested Schema for `metrics_endpoint`

Optional:

- `enabled` (Boolean) Set to `true` if metrics endpoint is enabled, `false` otherwise (default `false`).
- `source_ip_ranges` (List of String) IP addresses/blocks to allow traffic from when metrics endpoint is enabled.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.

## Deprovision / Destroy

By default, environments are protected against accidental deletion. The following attributes control the destroy behavior:

- **`force_destroy`** - Must be set to `true` and applied **before** running `terraform destroy`. Acts as a safety lock to prevent accidental deletion. Without a successful `terraform apply` after setting this flag, `terraform destroy` will fail.

- **`force_destroy_clusters`** - Set to `true` to automatically delete all provisioned clusters during the environment deletion. By default, `terraform destroy` will fail if there are active clusters in the environment.

- **`skip_deprovision_on_destroy`** - Set to `true` to skip the cloud resource cleanup and delete the environment record immediately. Useful when the environment was created with an immutable misconfiguration (e.g. wrong region) and no cloud infrastructure was actually provisioned. **Use with caution**: any provisioned resources are left behind.

- **`allow_delete_while_disconnected`** - Set to `true` to allow deletion when the environment is in a `DISCONNECTED` state. Commonly needed together with `skip_deprovision_on_destroy` when no infrastructure was created due to a configuration error; without this flag the destroy will always time out.

### Typical destroy workflow

1. Set `force_destroy = true` (and optionally `force_destroy_clusters = true`) in your configuration.
2. Run `terraform apply` to update the state.
3. Run `terraform destroy`.

```hcl
resource "altinitycloud_env_aws_hosted" "this" {
  # ... other configuration ...

  force_destroy          = true
  force_destroy_clusters = true
}
```
## Import

Import is supported using the following syntax:

```shell
terraform import altinitycloud_env_aws_hosted.this "replace-with-environment-name"
```
