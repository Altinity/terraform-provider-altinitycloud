---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "altinitycloud_env_hcloud Data Source - terraform-provider-altinitycloud"
subcategory: ""
description: |-
  Bring Your Own Cloud (BYOC) HCloud environment data source.
---

# altinitycloud_env_hcloud (Data Source)

Bring Your Own Cloud (BYOC) HCloud environment data source.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A globally-unique environment identifier. **[IMMUTABLE]**

		- All environment names must start with your account name as prefix.
		- ⚠️ Changing environment name after creation will force a resource replacement.

		Examples:
		- "acme-staging" (where "acme" is your account name)

### Read-Only

- `allow_delete_while_disconnected` (Boolean) Set to `true` to allow deletion of the environment while it is disconnected from the cloud connect. If the the environment is not connected during the deletion process you will end up in a delete timeout (default `false`).
- `cidr` (String) VPC CIDR block from the private IPv4 address ranges as specified in RFC 1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16). At least /21 required. **[IMMUTABLE]**

		Examples:
		- "10.136.0.0/21"
		- "172.20.0.0/21"
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
- `force_destroy` (Boolean) Locks the environment for accidental deletion when running `terraform destroy` command. Your environment will be deleted, only when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run (before running the `terraform destroy`) to update this value in the state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. (default `false`)
- `force_destroy_clusters` (Boolean) By default, the destroy operation will not delete any provisioned clusters and the deletion will fail until the clusters get removed. Set to `true` to remove all provisioned clusters as part of the environment deletion process.
- `hcloud_token_enc` (String) HCloud token (stored encrypted)
- `id` (String) ID of the environment (automatically generated based on the name)
- `load_balancers` (Attributes) Load balancers configuration. (see [below for nested schema](#nestedatt--load_balancers))
- `load_balancing_strategy` (String) Load balancing strategy for the environment.

		Possible Values:
		- "ROUND_ROBIN": load balance traffic across all zones in round-robin fashion (default)
		- "ZONE_BEST_EFFORT": keep traffic within same zone
- `locations` (List of String) Explicit list of HCloud locations. Currently supports single location only.

		Examples:
		- ["hil"]
- `maintenance_windows` (Attributes List) List of maintenance windows during which automatic maintenance is permitted. By default updates are applied as soon as they are available. (see [below for nested schema](#nestedatt--maintenance_windows))
- `network_zone` (String) HCloud network ([docs](https://docs.hetzner.com/cloud/general/locations)). **[IMMUTABLE]**

		Examples:
		- "us-west".
- `node_groups` (Attributes Set) List of node groups. At least one required. (see [below for nested schema](#nestedatt--node_groups))
- `skip_deprovision_on_destroy` (Boolean) Set to `true` will delete without waiting for environment deprovisioning. Use this with precaution, it may end up with dangling resources in your cloud provider (default `false`).
- `spec_revision` (Number) Spec revision
- `wireguard_peers` (Attributes List) HCloud Wireguard peer configuration. (see [below for nested schema](#nestedatt--wireguard_peers))

<a id="nestedatt--load_balancers"></a>
### Nested Schema for `load_balancers`

Optional:

- `internal` (Attributes) Internal load balancer configuration. Accessible via `*.internal.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--internal))
- `public` (Attributes) Public load balancer configuration. Accessible via `*.$env_name.altinity.cloud`. (see [below for nested schema](#nestedatt--load_balancers--public))

<a id="nestedatt--load_balancers--internal"></a>
### Nested Schema for `load_balancers.internal`

Optional:

- `enabled` (Boolean) Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)
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


<a id="nestedatt--node_groups"></a>
### Nested Schema for `node_groups`

Required:

- `capacity_per_location` (Number) Maximum number of instances per availability zone.
- `node_type` (String) List of node groups. At least one required.
- `reservations` (Set of String) Types of workload that are allowed to be scheduled onto the nodes that belong to this group.

		Possible values:
		- "SYSTEM" (at least one node group must include a SYSTEM reservation)
		- "CLICKHOUSE"
		- "ZOOKEEPER"

Optional:

- `locations` (List of String) Availability zones. Check possible available zones in your cloud provider documentation
- `name` (String) Unique (among environment node groups) node group identifier.


<a id="nestedatt--wireguard_peers"></a>
### Nested Schema for `wireguard_peers`

Required:

- `allowed_ips` (List of String) A list of addresses (in CIDR notation) that should get routed to the peer.
- `endpoint` (String) Peer endpoint.
- `public_key` (String) Peer public key.
