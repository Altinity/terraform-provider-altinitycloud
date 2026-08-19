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
