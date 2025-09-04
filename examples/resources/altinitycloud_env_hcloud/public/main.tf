resource "altinitycloud_env_certificate" "this" {
  env_name = "acme-staging"
}

variable "hcloud_token" {
  type = string
}

resource "altinitycloud_env_secret" "this" {
  pem   = altinitycloud_env_certificate.this.pem
  value = var.hcloud_token
}

resource "altinitycloud_env_hcloud" "this" {
  name             = altinitycloud_env_certificate.this.env_name
  cidr             = "10.136.0.0/21"
  network_zone     = "us-west"
  locations        = ["hil"]
  hcloud_token_enc = altinitycloud_env_secret.this.secret_value

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
      locations             = ["hil"]
    },
    {
      capacity_per_location = 10
      name                  = "ccx23"
      node_type             = "ccx23"
      reservations          = ["CLICKHOUSE"]
      locations             = ["hil"]
    }
  ]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_hcloud_status" "this" {
  name                           = altinitycloud_env_hcloud.this.name
  wait_for_applied_spec_revision = altinitycloud_env_hcloud.this.spec_revision
}
