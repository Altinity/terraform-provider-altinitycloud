resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

variable "datadog_api_key" {
  type      = string
  sensitive = true
}

// The Datadog API key is stored encrypted via the secret resource,
// the same pattern used for the Hetzner Cloud token.
resource "altinitycloud_env_secret" "datadog" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.datadog_api_key
}

provider "google" {
}

resource "google_project" "this" {
  project_id          = "YYYYYYYYYYYYYYYYYY"
  name                = "ZZZZZZZZZZZZZZZZZZ"
  auto_create_network = false
}

locals {
  zones = ["us-east1-b", "us-east1-d"]
}

resource "altinitycloud_env_gcp" "this" {
  name           = altinitycloud_env_certificate.this.env_name
  gcp_project_id = google_project.this.project_id
  region         = "us-east1"
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
      node_type         = "e2-standard-2"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "n2d-standard-2"
      capacity_per_zone = 10
      zones             = local.zones
      reservations      = ["CLICKHOUSE"]
    }
  ]

  datadog = {
    enabled         = true
    enc_api_key     = altinitycloud_env_secret.datadog.secret_value
    domain          = "datadoghq.com"
    logs_enabled    = true
    metrics_enabled = true
  }
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_gcp_status" "this" {
  name                           = altinitycloud_env_gcp.this.name
  wait_for_applied_spec_revision = altinitycloud_env_gcp.this.spec_revision
}
