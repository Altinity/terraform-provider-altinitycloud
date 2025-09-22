# Terraform Altinity.Cloud Provider

<div align="right">
  <img src="https://altinity.com/wp-content/uploads/2022/05/logo_horizontal_blue_white.svg" alt="Altinity" width="120">
</div>

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/altinity/altinitycloud/latest)
[![Latest Version](https://img.shields.io/badge/dynamic/json?label=version&query=$.version&url=https%3A//registry.terraform.io/v1/providers/altinity/altinitycloud)](https://registry.terraform.io/providers/altinity/altinitycloud/latest)
[![Documentation](https://img.shields.io/badge/-documentation-blue)](https://registry.terraform.io/providers/altinity/altinitycloud/latest/docs)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The official Terraform provider for [Altinity.Cloud](https://altinity.cloud/), enabling you to manage ClickHouse environments and infrastructure as code. This provider supports multiple cloud platforms including AWS, GCP, Azure, Hetzner Cloud, and Kubernetes clusters.

For detailed configuration options, see the [Terraform Registry documentation](https://registry.terraform.io/providers/altinity/altinitycloud/latest/docs).

If you're looking to set up the necessary infrastructure to connect your environments, see the connect modules:
- **AWS**: [terraform-altinitycloud-connect-aws](https://github.com/altinity/terraform-altinitycloud-connect-aws)
- **Kubernetes**: [terraform-altinitycloud-connect](https://github.com/altinity/terraform-altinitycloud-connect)

## Prerequisites

Before using this provider, ensure you have:

1. **Terraform** >= 1.0
2. **Altinity.Cloud account** with Anywhere API access
3. **Cloud provider credentials** (AWS, GCP, Azure, etc.) for the target environment

## Usage

### Basic Setup

```terraform
terraform {
  required_providers {
    altinitycloud = {
      source  = "altinity/altinitycloud"
      version = "~> 0.4.0"
    }
  }
}

provider "altinitycloud" {
  # API token can be set via ALTINITYCLOUD_API_TOKEN environment variable
  api_token = "your-api-token-here"
}
```

## Examples

For comprehensive examples covering all supported cloud platforms and use cases, see the [examples directory](examples/):

- **Resources**: AWS, Azure, GCP, Hetzner Cloud, and Kubernetes environments with various configurations
- **Data Sources**: Environment status monitoring and validation examples
- **Advanced Configurations**: VPC peering, custom networking, secrets management, and more

Quick example for AWS:

```terraform
resource "altinitycloud_env_certificate" "example" {
  env_name = "my-clickhouse-env"
}

resource "altinitycloud_env_aws" "example" {
  name           = altinitycloud_env_certificate.example.env_name
  aws_account_id = "123456789012"
  region         = "us-west-2"
  zones          = ["us-west-2a", "us-west-2b"]
  cidr           = "10.67.0.0/21"

  node_groups = [
    {
      node_type         = "m6i.large"
      capacity_per_zone = 5
      reservations      = ["CLICKHOUSE"]
    }
  ]
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `altinitycloud_env_certificate` | Generate environment certificates for cloud connection |
| `altinitycloud_env_aws` | Manage AWS-based ClickHouse environments |
| `altinitycloud_env_azure` | Manage Azure-based ClickHouse environments |
| `altinitycloud_env_gcp` | Manage GCP-based ClickHouse environments |
| `altinitycloud_env_hcloud` | Manage Hetzner Cloud-based ClickHouse environments |
| `altinitycloud_env_k8s` | Manage Kubernetes-based ClickHouse environments |
| `altinitycloud_env_secret` | Manage environment secrets and configuration |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `altinitycloud_env_aws_status` | Monitor AWS environment provisioning status |
| `altinitycloud_env_azure_status` | Monitor Azure environment provisioning status |
| `altinitycloud_env_gcp_status` | Monitor GCP environment provisioning status |
| `altinitycloud_env_hcloud_status` | Monitor Hetzner Cloud environment provisioning status |
| `altinitycloud_env_k8s_status` | Monitor Kubernetes environment provisioning status |

## Troubleshooting

- **Authentication errors:** Ensure your API token is valid and has proper permissions. Check the `ALTINITYCLOUD_API_TOKEN` environment variable.
- **Environment provisioning fails:** Use the corresponding status data source to wait for provisioning completion and get detailed error information.
- **Resource dependencies:** Always use status data sources to ensure proper resource ordering and error detection.

### Need Help?

If you encounter issues not covered above, please [create an issue](https://github.com/altinity/terraform-provider-altinitycloud/issues/new) with detailed information about your problem.

## Contributing

Contributions are welcome! Please submit a Pull Request or open an issue for major changes. See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for development guidelines and setup instructions.

## License

All code, unless specified otherwise, is licensed under the [Apache-2.0](LICENSE) license.
Copyright (c) 2023 Altinity, Inc.
