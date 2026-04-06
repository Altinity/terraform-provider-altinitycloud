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

### Invalid or missing API token

```
Missing Altinity.Cloud API Token
ALTINITYCLOUD_API_TOKEN environment variable or "api_token" provider attribute required.
```

Make sure you have set the `ALTINITYCLOUD_API_TOKEN` environment variable or the `api_token` provider attribute. If the token is set but you get an `invalid API token` error, generate a new one from [ACM](https://acm.altinity.cloud/) under "My Account" > "Anywhere API Access".

### Immutable attribute modified

```
Immutable Attribute
{attribute_name} is immutable and cannot be modified after creation.
```

Some attributes like `name`, `region`, and `cidr` cannot be changed after the environment is created. To change them, you need to destroy the environment and create a new one.

### Environment provisioning errors

During provisioning, the environment status may report errors like:

- **`DISCONNECTED`** / **`K8S_DISCONNECTED`** — The environment is not connected to Altinity.Cloud. Verify that the cloud connect setup was completed correctly.
- **`CLOUD_PROVIDER_ACCESS_DENIED`** — The credentials or permissions provided to Altinity.Cloud are insufficient. Check IAM roles, service accounts, or service principals depending on your cloud provider.
- **`CLOUD_PROVIDER_QUOTA_EXCEEDED`** — Your cloud provider account has reached a resource quota limit. Request a quota increase from your cloud provider.
- **`CLOUD_PROVIDER_RESOURCE_NOT_FOUND`** / **`GCP_PROJECT_NOT_FOUND`** — A referenced cloud resource (e.g., project, subscription, account) does not exist. Verify the resource identifiers in your configuration.

These errors are visible in the `altinitycloud_env_*_status` data source and in the [ACM](https://acm.altinity.cloud/) UI.

### MFA timeout during destroy

```
timeout reached while waiting for MFA to be confirmed.
Please check your MFA device, confirm deletion and run `terraform destroy` again
```

If your organization has MFA enabled, environment deletion requires manual approval. After running `terraform destroy`, confirm the deletion on your MFA device within 5 minutes. If the timeout is reached, simply run `terraform destroy` again.

### Destroying an environment

Environment deletion is protected by default and requires specific flags to be set before running `terraform destroy`. Common issues include: the environment being locked (`force_destroy`), having active clusters (`force_destroy_clusters`), or being disconnected (`allow_delete_while_disconnected`).

See the **Deprovision / Destroy** section in each environment resource documentation for detailed instructions.

## Support

If you need help, reach out to us via Slack:

- **Enterprise customers**: Use your organization's dedicated Altinity Slack channel.
- **Community**: Join the [AltinityDB workspace](https://altinitydbworkspace.slack.com/) and post in the **#terraform** channel.

You can also [create an issue](https://github.com/altinity/terraform-provider-altinitycloud/issues/new) on GitHub.

## Contributing

Contributions are welcome! Please submit a Pull Request or open an issue for major changes. See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for development guidelines and setup instructions.

## License

All code, unless specified otherwise, is licensed under the [Apache-2.0](LICENSE) license.
Copyright (c) 2023 Altinity, Inc.
