---
page_title: "ClickHouse clusters"
subcategory: ""
description: |-
  Provision ClickHouse clusters and Keepers inside an Altinity.Cloud environment.
---

# ClickHouse clusters

ClickHouse clusters and Keepers are not standalone resources. They are nested list attributes, `clickhouse_clusters` and `clickhouse_keepers`, on the environment resource. Adding an entry to a list creates it, editing an entry updates it, and removing an entry deletes it.

~> **Note:** Only `altinitycloud_env_aws` supports `clickhouse_clusters` and `clickhouse_keepers` today. Other environment types will follow.

How the pieces fit together:

- The environment provides **node groups**. Each node group has `reservations` that say what it may run.
- A Keeper needs a node group with a `ZOOKEEPER` reservation whose `node_type` matches the Keeper's `instance_type`.
- A cluster needs a node group with a `CLICKHOUSE` reservation whose `node_type` matches the cluster's `instance_type`.
- A Keeper is only needed for replication: a cluster with `replicas` greater than 1 must set `keeper` to the name of one of the entries in `clickhouse_keepers`.

For the complete attribute reference, see [`clickhouse_clusters`](../resources/env_aws#nestedatt--clickhouse_clusters) and `clickhouse_keepers` in the `altinitycloud_env_aws` docs.

## Prerequisites

- A working AWS environment. See [Basic environment](../resources/env_aws#basic-environment).
- An `altinitycloud_env_certificate` and an `altinitycloud_env_secret`. User password digests are encrypted with the environment certificate before they are sent.

## Minimal cluster

A cluster only needs three attributes, the same fields the ACM "Launch a Cluster" dialog asks for:

- `name`
- `instance_type`
- `disk.size_in_gib`

Everything else has a default: 1 shard, 1 replica, no Keeper, the platform's current ClickHouse image, and the first three environment zones.

```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

locals {
  zones = ["us-east-1a", "us-east-1b"]
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
  zones          = local.zones
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
      zones             = local.zones
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["CLICKHOUSE"]
    }
  ]

  // Single shard, single replica, default image. Add users to be able to connect.
  clickhouse_clusters = [
    {
      name          = "hello"
      instance_type = "m6i.large" // must match a node group with the CLICKHOUSE reservation
      disk = {
        size_in_gib = 100
      }
    }
  ]

  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
```

~> **Note:** No ClickHouse users are created by default, so nobody can log in to this cluster yet. Add a user, as shown below, to connect.

## Hello world

The minimal cluster plus a user you can connect with:

```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

// SHA256 hex digest of the password, e.g. `echo -n 'secret' | sha256sum`.
variable "admin_password_sha256_hex" {
  type      = string
  sensitive = true
}

resource "altinitycloud_env_secret" "admin_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.admin_password_sha256_hex
}

locals {
  zones = ["us-east-1a", "us-east-1b"]
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
  zones          = local.zones
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
      zones             = local.zones
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["CLICKHOUSE"]
    }
  ]

  clickhouse_clusters = [
    {
      name          = "hello"
      instance_type = "m6i.large" // must match a node group with the CLICKHOUSE reservation
      disk = {
        size_in_gib = 100
      }
      users = [
        {
          name = "admin"
          // Without networks.ips the user can only connect from localhost.
          networks = {
            ips = ["0.0.0.0/0"]
          }
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.admin_password.secret_value
          }
        }
      ]
    }
  ]

  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// Waits for the environment and cluster to be reconciled and reports provisioning errors.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}

output "clickhouse_endpoint" {
  value = "https://hello.${altinitycloud_env_aws.this.name}.altinity.cloud:8443"
}
```

Things to note:

- The password is the SHA256 hex digest of the real password (`echo -n 'secret' | sha256sum`), encrypted through `altinitycloud_env_secret`. The plain password never appears in the configuration.
- When `networks.ips` is omitted or empty, the user can only connect from localhost (`127.0.0.1` and `::1`). List CIDRs to open it up.
- Provisioning is asynchronous. `altinitycloud_env_aws_status` with `wait_for_applied_spec_revision` waits for the environment to be reconciled and surfaces provisioning errors. Without it, Terraform cannot detect failures.
- The cluster endpoint, `https://<cluster>.<env>.altinity.cloud:8443`, comes up shortly after the data source returns.

## Day-2 operations

The snippets below show only the relevant fragment of the environment resource.

### Scale replicas and shards

Going above one replica requires a Keeper. See [Highly available Keeper](#highly-available-keeper).

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name     = "hello"
      shards   = 2
      replicas = 3
      # ...
    }
  ]
}
```

### Grow a disk

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name = "hello"
      # ...
      disk = {
        size_in_gib = 100
      }
    }
  ]
}
```

~> **Note:** Disk size can only be increased, never decreased.

### Stop and start

Set `stopped = true` to keep a cluster stopped, and set it back to `false` to start it again. Keepers have the same attribute.

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name    = "hello"
      stopped = true
      # ...
    }
  ]
}
```

### Upgrade

Change `image`, tag included. When `image` is omitted, the platform picks a current image.

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name  = "hello"
      image = "altinity/clickhouse-server:25.8.16.10002.altinitystable"
      # ...
    }
  ]
}
```

### Add a user

`password.enc_value` is required when a user is created. The API never returns it, so Terraform keeps the value from your configuration and never refreshes it.

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name = "hello"
      # ...
      users = [
        {
          name = "admin"
          # ...
        },
        {
          name = "reporting"
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.reporting_password.secret_value
          }
        }
      ]
    }
  ]
}
```

Once the user exists, `password.enc_value` can be omitted to keep the current password.

### Rotate a password

Create a new `altinitycloud_env_secret` from the new digest and point the user's `enc_value` at it.

### Remove a cluster

Delete its entry from `clickhouse_clusters`. This deletes the cluster and its data.

~> **Warning:** Changing a cluster's or Keeper's `name` is **not** a rename. The old cluster is deleted along with its data and a new, empty one is created.

~> **Warning:** `zones` is immutable on both clusters and Keepers.

~> **Warning:** A Keeper ensemble of 3 cannot be shrunk back to a single node.

## Ownership and import

Terraform only manages the clusters and Keepers declared in the configuration:

- Clusters and Keepers created elsewhere, for example in the ACM console, are left alone.
- Declaring a name that already exists in the environment is rejected rather than overwritten.
- To bring existing clusters under Terraform, import the environment. This adopts every existing cluster and Keeper into state:

```shell
terraform import altinitycloud_env_aws.this "<env-name>"
```

After the import, make your configuration match what was adopted and run `terraform plan` to review any differences.

## Advanced configuration

### Highly available Keeper

A replicated cluster coordinates through a Keeper. The Keeper's `instance_type` must match a node group with the `ZOOKEEPER` reservation. Use three Keeper replicas for a highly available ensemble.

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_keepers = [
    {
      name          = "keeper"
      instance_type = "t4g.large"
      replicas      = 3
      disk = {
        size_in_gib = 30
      }
    }
  ]
  clickhouse_clusters = [
    {
      name     = "hello"
      replicas = 2
      keeper   = "keeper"
      # ...
    }
  ]
}
```

### Extra disks and disk performance

`extra_disks` adds data volumes to the default ClickHouse storage policy. Each replica receives one volume per entry, and ClickHouse balances new data across the main and extra volumes with the `least_used` strategy.

- At most 8 extra disks.
- `name` is at most 16 characters, must start with `disk`, and is immutable.
- A volume named `diskN`, where `N` is a positive integer, is mounted at `/var/lib/clickhouseN`.

`iops` and `throughput_in_mibps` are honored only by storage classes that support them (`iops`: gp3 and io2; `throughput_in_mibps`: gp3).

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name = "analytics"
      # ...
      disk = {
        size_in_gib         = 100
        iops                = 6000
        throughput_in_mibps = 250
      }
      extra_disks = [
        {
          name        = "disk1"
          size_in_gib = 500
        }
      ]
    }
  ]
}
```

### Server settings

Server-level settings applied to every node. See the [ClickHouse server settings reference](https://clickhouse.com/docs/reference/settings/server-settings/settings).

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name = "analytics"
      # ...
      server_settings = [
        {
          key   = "max_concurrent_queries"
          value = "200"
        }
      ]
    }
  ]
}
```

### Profiles and quotas

Profiles apply session-level settings (see the [session settings reference](https://clickhouse.com/docs/reference/settings/session-settings)). Quotas limit resource usage over an interval; exactly one interval is currently supported. Users reference them by name with `profile` and `quota`.

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_clusters = [
    {
      name = "analytics"
      # ...
      profiles = [
        {
          name = "limited"
          settings = [
            {
              key   = "max_execution_time"
              value = "30"
            }
          ]
        }
      ]
      quotas = [
        {
          name = "hourly"
          intervals = [
            {
              duration = 3600
              queries  = 1000
            }
          ]
        }
      ]
      users = [
        {
          name    = "app"
          profile = "limited"
          quota   = "hourly"
          # ...
        }
      ]
    }
  ]
}
```

### User permissions

By default a user has access to all databases. Narrow it with `databases`, or use `grants` for fine-grained control:

```terraform
users = [
  {
    name      = "app"
    databases = ["analytics"]
    # ...
  },
  {
    name   = "readonly"
    grants = ["GRANT SELECT ON analytics.*"]
    # ...
  }
]
```

`grants` cannot be combined with `databases`, `access_management`, `named_collection_control` or `show_named_collections_secrets`.

The user names `grafana` and `datadog` are reserved for platform-injected users and are rejected.

### Multiple clusters sharing a Keeper

Several clusters can point at the same Keeper by name:

```terraform
resource "altinitycloud_env_aws" "this" {
  # ...
  clickhouse_keepers = [
    {
      name          = "keeper"
      instance_type = "t4g.large"
      replicas      = 3
      # ...
    }
  ]
  clickhouse_clusters = [
    {
      name   = "analytics"
      keeper = "keeper"
      # ...
    },
    {
      name   = "events"
      keeper = "keeper"
      # ...
    }
  ]
}
```

## Full example

A highly available Keeper and a replicated cluster with an extra disk, server settings, two users, a profile and a quota:

```terraform
resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

// SHA256 hex digests of each user's password, stored encrypted via the secret resource.
// A password is required to create a user; it can be omitted later to keep the current one.
variable "app_password_sha256_hex" {
  type      = string
  sensitive = true
}

variable "readonly_password_sha256_hex" {
  type      = string
  sensitive = true
}

resource "altinitycloud_env_secret" "app_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.app_password_sha256_hex
}

resource "altinitycloud_env_secret" "readonly_password" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.readonly_password_sha256_hex
}

locals {
  zones = ["us-east-1a", "us-east-1b"]
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
  zones          = local.zones
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
      zones             = local.zones
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "m6i.large"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["CLICKHOUSE"]
    }
  ]

  clickhouse_keepers = [
    {
      name          = "keeper"
      instance_type = "t4g.large"
      zones         = local.zones
      replicas      = 3
      disk = {
        size_in_gib = 30
      }
    }
  ]

  clickhouse_clusters = [
    {
      name          = "analytics"
      image         = "altinity/clickhouse-server:24.8.14.10459.altinitystable" // omit to get the platform default
      instance_type = "m6i.large"
      zones         = local.zones
      shards        = 1
      replicas      = 2
      keeper        = "keeper" // must name one of clickhouse_keepers
      disk = {
        size_in_gib = 100
      }
      // Extra volumes join the default storage policy; names must start with "disk".
      extra_disks = [
        {
          name        = "diskcold"
          size_in_gib = 500
        }
      ]
      server_settings = [
        {
          key   = "max_concurrent_queries"
          value = "200"
        }
      ]
      users = [
        {
          name    = "app"
          profile = "default"
          quota   = "default"
          networks = {
            ips = ["10.0.0.0/8"]
          }
          databases = ["analytics"]
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.app_password.secret_value
          }
        },
        {
          // "grants" cannot be combined with "databases" or the access flags.
          name   = "readonly"
          grants = ["GRANT SELECT ON analytics.*"]
          password = {
            type      = "SHA256_HEX"
            enc_value = altinitycloud_env_secret.readonly_password.secret_value
          }
        }
      ]
      profiles = [
        {
          name = "default"
          settings = [
            {
              key   = "max_execution_time"
              value = "30"
            }
          ]
        }
      ]
      quotas = [
        {
          name = "default"
          intervals = [
            {
              duration = 3600
              queries  = 1000
            }
          ]
        }
      ]
    }
  ]

  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
// Cluster endpoints (https://<cluster>.<env>.altinity.cloud:8443) come up shortly after it returns.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
```

## Destroy

Destroying the environment follows the usual flow described in [Deprovision / Destroy](../resources/env_aws#deprovision--destroy). By default `terraform destroy` fails while clusters are still running; set `force_destroy_clusters = true` (together with `force_destroy = true`) and apply before destroying to delete them with the environment.
