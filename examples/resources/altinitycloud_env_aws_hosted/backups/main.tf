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
