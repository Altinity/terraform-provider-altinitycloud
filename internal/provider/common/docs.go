package common

// Shared descriptions.
const ID_DESCRIPTION = "ID of the environment (automatically generated based on the name)"
const NAME_DESCRIPTION = `A globally-unique environment identifier. All environment names must start with your account name as prefix. **[IMMUTABLE]**

		Examples:
		- "acme-staging" (where "acme" is your account name)
`
const CIDR_DESCRIPTION = `VPC CIDR block from the private IPv4 address ranges as specified in RFC 1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16). At least /21 required. **[IMMUTABLE]**

		Examples:
		- "10.136.0.0/21"
		- "172.20.0.0/21"
`
const CUSTOM_DOMAIN_DESCRIPTION = `Custom domain.

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
`
const NUMBER_OF_ZONES_DESCRIPTION = "Number of zones where the environment will be available. When set, zones will be set automatically based on your cloud provider (Do not use it together with zones)"
const SOURCE_IP_RANGES_DESCRIPTION = " IP addresses/blocks to allow traffic from (default `\"0.0.0.0/0\"`)."
const MAINTENANCE_WINDOW_DAYS_DESCRIPTION = `Days on which maintenance can take place.

		Possible values:
		- "MONDAY"
		- "TUESDAY"
		- "WEDNESDAY"
		- "THURSDAY"
		- "FRIDAY"
		- "SATURDAY"
		- "SUNDAY"
`
const MAINTENANCE_WINDOW_NAME_DESCRIPTION = "Maintenance window identifier"
const MAINTENANCE_WINDOW_HOUR_DESCRIPTION = "Hour of the day in [0, 23] range."
const MAINTENANCE_WINDOW_LENGTH_IN_HOURS_DESCRIPTION = "Maintenance window length in hours. 4h min, 24h max."
const MAINTENANCE_WINDOW_ENABLED_DESCRIPTION = "Set to `true` if maintenance window is enabled, `false` otherwise. (default `false`)"
const MAINTENANCE_WINDOW_DESCRIPTION = "List of maintenance windows during which automatic maintenance is permitted. By default updates are applied as soon as they are available."
const KEY_DESCRIPTION = "Name of the key"
const VALUE_DESCRIPTION = "Value of the key"
const LOAD_BALANCING_STRATEGY_DESCRIPTION = `Load balancing strategy for the environment.

		Possible Values:
		- "ROUND_ROBIN": load balance traffic across all zones in round-robin fashion (default)
		- "ZONE_BEST_EFFORT": keep traffic within same zone
`
const LOAD_BALANCER_PUBLIC_DESCRIPTION = "Public load balancer configuration. Accessible via `*.$env_name.altinity.cloud`."
const LOAD_BALANCER_INTERNAL_DESCRIPTION = "Internal load balancer configuration. Accessible via `*.internal.$env_name.altinity.cloud`."
const LOAD_BALANCER_DESCRIPTION = "Load balancers configuration."
const LOAD_BALANCER_ENABLED_DESCRIPTION = "Set to `true` if load balancer is enabled, `false` otherwise. (default `false`)"
const NODE_GROUP_DESCRIPTION = "List of node groups. At least one required."
const NODE_GROUP_CAPACITY_PER_ZONE_DESCRIPTION = "Maximum number of instances per availability zone."
const NODE_GROUP_ZONES_DESCRIPTION = "Availability zones. Check possible available zones in your cloud provider documentation"
const NODE_GROUP_RESERVATIONS_DESCRIPTION = `Types of workload that are allowed to be scheduled onto the nodes that belong to this group.

		Possible values:
		- "SYSTEM" (at least one node group must include a SYSTEM reservation)
		- "CLICKHOUSE"
		- "ZOOKEEPER"
`
const NODE_GROUP_SELECTOR_DESCRIPTION = "`nodeSelector` to apply to the pods targeting this group"
const NODE_GROUP_NAME_DESCRIPTION = "Unique (among environment node groups) node group identifier."
const NODE_GROUP_TOLERATIONS = "List of tolerations to apply to the pods targeting this group"
const NODE_GROUP_TOLERATIONS_KEY = "Taint key, e.g. 'dedicated'"
const NODE_GROUP_TOLERATIONS_VALUE = "Taint value, e.g. 'clickhouse'"
const NODE_GROUP_TOLERATIONS_EFFECT = `Node taint effect.

		Possible values:
		- "NO_SCHEDULE"
		- "PREFER_NO_SCHEDULE"
		- "NO_EXECUTE"
`
const NODE_GROUP_TOLERATIONS_OPERATOR = `Node toleration operator used to match taints.

		Possible values:
		- "EQUALS"
		- "EXISTS"
`

const FORCE_DESTROY_DESCRIPTION = "Locks the environment for accidental deletion when running `terraform destroy` command. Your environment will be deleted, only when setting this parameter to `true`. Once this parameter is set to `true`, there must be a successful `terraform apply` run (before running the `terraform destroy`) to update this value in the state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. (default `false`)"
const FORCE_DESTROY_CLUSTERS_DESCRIPTION = "By default, the destroy operation will not delete any provisioned clusters and the deletion will fail until the clusters get removed. Set to `true` to remove all provisioned clusters as part of the environment deletion process."
const SKIP_PROVISIONING_ON_DESTROY_DESCRIPTION = "Set to `true` will delete without waiting for environment deprovisioning. Use this with precaution (default `false`)."
const STATUS_DESCRIPTION = "Environment status"
const STATUS_SPEC_REVISION_DESCRIPTION = "Spec revision"
const STATUS_APPLIED_SPEC_REVISION_DESCRIPTION = "Applied spec revision"
const STATUS_PENDING_DELETE_DESCRIPTION = "`true` indicates that environment is pending deletion"
const STATUS_LOAD_BALANCERS_DESCRIPTION = "Status of internal load balancer."
const STATUS_LOAD_BALANCERS_INTERNAL_DESCRIPTION = "Status of load balancers."
const STATUS_LOAD_BALANCERS_ENDPOINT_SERVICE_NAME_DESCRIPTION = "VPC endpoint service name in $endpoint_service_id.$region.vpce.amazonaws.com format (if any)"

// AWS descriptions.
const AWS_ACCOUNT_ID_DESCRIPTION = "ID of the AWS account ([docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html#ViewYourAWSId)) in which to provision AWS resources. **[IMMUTABLE]**"
const AWS_TAGS_DESCRIPTION = "Tags to apply to AWS resources."
const AWS_REGION_DESCRIPTION = `AWS region ([docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html#Concepts.RegionsAndAvailabilityZones.Regions)). **[IMMUTABLE]**

		Examples:
		- "us-east-1"
		- "sa-east-1"
`
const AWS_ZONES_DESCRIPTION = `Explicit list of AWS availability zones. At least 2 required.

		Examples:
		- ["us-east-1a", "us-east-1b"]
		- ["sa-east-1c", "sa-east-1d"]
`
const AWS_LOAD_BALANCER_CROSS_ZONE_DESCRIPTION = "`true` indicates that traffic should be distributed across all specified availability zones, `false` otherwise. (default `true`)."
const AWS_LOAD_BALANCER_ENDPOINT_SERVICE_ALLOWED_PRINCIPALS_DESCRIPTION = `ARNs for AWS principals that are allowed to create VPC endpoints.

		Examples:
		- "arn:aws:iam::$account_id:root"
`
const AWS_NODE_GROUP_NODE_TYPE_DESCRIPTION = `Instance type ([docs](https://aws.amazon.com/ec2/instance-types/))

		Examples:
		- "t4g.large"
`
const PEERING_CONNECTION_DESCRIPTION = "AWS environment VPC peering configuration."
const PEERING_CONNECTION_ID_DESCRIPTION = "VPC peering connection ID."
const PEERING_CONNECTION_VPC_ID_DESCRIPTION = "Target VPC ID."
const PEERING_CONNECTION_VPC_REGION_DESCRIPTION = "Target VPC region (defaults to environment region)."
const PEERING_CONNECTION_AWS_ACCOUNT_ID_DESCRIPTION = "Target VPC AWS account ID (defaults to environment AWS account ID)."
const ENDPOINT_DESCRIPTION = "AWS environment VPC endpoint configuration"
const ENDPOINT_SERVICE_NAME_DESCRIPTION = "VPC endpoint service name in $endpoint_service_id.$region.vpce.amazonaws.com format."
const ENDPOINT_ALIAS_DESCRIPTION = "By default, VPC endpoints get assigned $endpoint_service_id.$env_name.altinity.cloud DNS record. Alias allows to override DNS record name to `$alias.$env_name.altinity.cloud`."
const CLOUD_CONNECT_DESCRIPTION = "`true` indicates that cloud resources are to be managed via altinity/cloud-connect and `false` means direct management (default `true`). **[IMMUTABLE]**"

// GCP descriptions.
const GCP_REGION_DESCRIPTION = `GCP region ([docs](https://cloud.google.com/about/locations)). **[IMMUTABLE]**

		Examples:
		- "us-west1".
`
const GCP_ZONES_DESCRIPTION = `Explicit list of GCP zones. At least 2 required.
		Examples:
		- ["us-west1a", "us-west1b"]
`
const GCP_NODE_GROUP_NODE_TYPE_DESCRIPTION = `Machine type ([docs](https://cloud.google.com/compute/docs/machine-resource)).

		Examples:
		- "e2-standard-2"
`
const GCP_PROJECT_ID_DESCRIPTION = "ID of the GCP project ([docs](https://support.google.com/googleapi/answer/7014113?hl=en#:~:text=The%20project%20ID%20is%20a,ID%20or%20create%20your%20own.)) in which to provision GCP resources. **[IMMUTABLE]**"

// K8S descriptions.
const K8S_NODE_GROUP_NODE_TYPE_DESCRIPTION = "node.kubernetes.io/instance-type value."
const K8S_NODE_GROUP_ZONES_DESCRIPTION = "topology.kubernetes.io/zone values."
const K8S_REGION_DESCRIPTION = "Cloud provider Region. Check possible available regions in your cloud provider documentation **[IMMUTABLE]**"
const K8S_LOAD_BALANCER_ANNOTATIONS_DESCRIPTION = "List of annotations for the load balancer"
const DISTRIBUTION_DESCRIPTION = `Kubernetes distribution. **[IMMUTABLE]**

		Possible values:
		- "AKS"
		- "EKS"
		- "GKE"
		- "CUSTOM"
`
const CUSTOM_NODE_TYPES_DESCRIPTION = "Custom node types"
const CUSTOM_NODE_TYPES_NAME_DESCRIPTION = "Custom node type unique identifier"
const CUSTOM_NODE_TYPES_CPU_ALLOCATABLE_DESCRIPTION = "Number of allocatable virtual cores"
const CUSTOM_NODE_TYPES_MEMORY_ALLOCATABLE__DESCRIPTION = "Amount of allocatable memory in bytes"
const LOGS_DESCRIPTION = "Kubernetes environment logs configuration"
const STORAGE_DESCRIPTION = "Storage backend configuration"
const S3_STORAGE_DESCRIPTION = "Amazon S3 configuration"
const GCS_STORAGE_DESCRIPTION = "Google Cloud Storage configuration"
const BUCKET_NAME_DESCRIPTION = "Bucket name"
const METRICS_DESCRIPTION = "Metrics configuration"
const METRICS_RETENTION_PERIOD_IN_DAYS_DESCRIPTION = "Metrics retention period in days (default `30`)."

// Azure descriptions.
const AZURE_CUSTOM_DOMAIN_DESCRIPTION = `Custom domain.

		Examples:
		- "example.com"
		- "foo.bar.com"

		Before specifying custom domain, please create the following DNS records:
		- CNAME _acme-challenge.example.com. $env_name.altinity.cloud.
		- (optional, public load balancer)
			CNAME *.example.com. _.$env_name.altinity.cloud.
		- (optional, internal load balancer)
			CNAME *.internal.example.com. _.internal.$env_name.altinity.cloud.
		- (optional, privatelink)
			CNAME *.privatelink.example.com. _.privatelink.$env_name.altinity.cloud.
`
const AZURE_ZONES_DESCRIPTION = `Explicit list of Azure availability zones. At least 2 required.

		Examples:
		- ["eastus-1", "eastus-2"]
`
const AZURE_REGION_DESCRIPTION = `Azure region ([docs](https://azure.microsoft.com/en-us/explore/global-infrastructure/geographies/#overview)). **[IMMUTABLE]**

		Examples:
		- "eastus"
		- "westus"
`

const AZURE_TENANT_ID_DESCRIPTION = "ID of the Azure Active Directory tenant for user identity and access management. **[IMMUTABLE]**"
const AZURE_SUBSCRIPTION_ID_DESCRIPTION = "ID linking the environment to a specific Azure subscription for resource management. **[IMMUTABLE]**"
const AZURE_PRIVATE_LINK_SERVICE_DESCRIPTION = "Azure Private Link service configuration."
const AZURE_PRIVATE_LINK_SERVICE_ALIAS_DESCRIPTION = "Private Link Service Alias / DNS Name in prefix.GUID.suffix format."
const AZURE_PRIVATE_LINK_SERVICE_ALLOWED_SUBSCRIPTIONS_DESCRIPTION = "Lists subscription IDs permitted for Private Link access, securing service connections."
const AZURE_TAGS_DESCRIPTION = "Tags to apply to Azure resources."
