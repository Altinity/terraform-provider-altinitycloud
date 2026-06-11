resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

locals {
  zones = ["us-east-1a", "us-east-1b"]

  // Stable prefix for the IAM roles the environment provisions. The KMS key
  // policy references these roles by prefix BEFORE the env exists, so it must
  // be set explicitly (not auto-generated) when bringing your own key.
  resource_prefix = "acme-staging"

  // Altinity's PRODUCTION AWS Organization ID. Dev/staging use a different org
  // — the env's "Encryption" page in the console renders whichever applies.
  altinity_org_id = "o-u95tkx5okz"
}

provider "aws" {
  region = "us-east-1"
}

data "aws_caller_identity" "current" {}

module "altinitycloud_connect_aws" {
  source = "altinity/connect-aws/altinitycloud"
  pem    = altinitycloud_env_certificate.this.pem
}

// Customer-managed KMS key (SaaS form). Encrypts the env's EBS volumes, EKS
// root volumes, and all reconciler-created S3 buckets.
resource "aws_kms_key" "altinity_env" {
  description             = "Altinity ${local.resource_prefix} environment encryption key"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        // Lets this env's reconciler roles use the key and create grants on it.
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

resource "altinitycloud_env_aws" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  aws_account_id = "123456789012"
  region         = "us-east-1"
  zones          = local.zones
  cidr           = "10.67.0.0/21"

  // Pins the IAM role names so the KMS policy above can target them by prefix
  // before the env is provisioned. Requires a permissions boundary to be set.
  resource_prefix                 = local.resource_prefix
  permissions_boundary_policy_arn = "arn:aws:iam::123456789012:policy/altinity-boundary"

  // Env-level customer-managed KMS key: encrypts all Altinity-provisioned
  // data buckets and EBS volumes. Immutable — set at creation only.
  kms_key_arn = aws_kms_key.altinity_env.arn

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
  cloud_connect = true
  depends_on = [
    // "depends_on" is here to enforce "this resource, then altinitycloud_connect_aws" order on destroy.
    module.altinitycloud_connect_aws
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_aws_status" "this" {
  name                           = altinitycloud_env_aws.this.name
  wait_for_applied_spec_revision = altinitycloud_env_aws.this.spec_revision
}
