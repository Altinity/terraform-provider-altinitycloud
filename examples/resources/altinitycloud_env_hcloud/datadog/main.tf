resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

variable "hcloud_token" {
  type      = string
  sensitive = true
}

variable "datadog_api_key" {
  type      = string
  sensitive = true
}

resource "altinitycloud_env_secret" "token" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.hcloud_token
}

// The Datadog API key is stored encrypted via the secret resource,
// the same pattern used for the Hetzner Cloud token above.
resource "altinitycloud_env_secret" "datadog" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.datadog_api_key
}

locals {
  locations = ["hil"]
}

resource "altinitycloud_env_hcloud" "this" {
  name             = altinitycloud_env_certificate.this.env_name
  cidr             = "10.136.0.0/21"
  network_zone     = "us-west"
  locations        = local.locations
  hcloud_token_enc = altinitycloud_env_secret.token.secret_value

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }

  node_groups = [
    {
      capacity_per_location = 10
      name                  = "cpx11"
      node_type             = "cpx11"
      reservations          = ["SYSTEM", "ZOOKEEPER"]
      locations             = local.locations
    },
    {
      capacity_per_location = 10
      name                  = "ccx23"
      node_type             = "ccx23"
      reservations          = ["CLICKHOUSE"]
      locations             = local.locations
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
data "altinitycloud_env_hcloud_status" "this" {
  name                           = altinitycloud_env_hcloud.this.name
  wait_for_applied_spec_revision = altinitycloud_env_hcloud.this.spec_revision
}
