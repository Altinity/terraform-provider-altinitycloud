terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
    altinitycloud = {
      source = "altinity/altinitycloud"
    }
  }
}

provider "google" {
}

resource "google_project" "this" {
  project_id          = "YYYYYYYYYYYYYYYYYY"
  name                = "ZZZZZZZZZZZZZZZZZZ"
  auto_create_network = false
}

resource "google_project_iam_member" "this" {
  for_each = toset([
    # https://cloud.google.com/iam/docs/understanding-roles
    "roles/compute.admin",
    "roles/container.admin",
    "roles/dns.admin",
    "roles/storage.admin",
    "roles/storage.hmacKeyAdmin",
    "roles/iam.serviceAccountAdmin",
    "roles/iam.serviceAccountKeyAdmin",
    "roles/iam.serviceAccountTokenCreator",
    "roles/iam.serviceAccountUser",
    "roles/iam.workloadIdentityPoolAdmin",
    "roles/serviceusage.serviceUsageAdmin",
    "roles/resourcemanager.projectIamAdmin",
    "roles/iap.tunnelResourceAccessor"
  ])
  project = google_project.this.id
  role    = each.key
  member  = "group:anywhere-admin@altinity.com"
}

resource "altinitycloud_env_gcp" "this" {
  name           = "acme-staging"
  gcp_project_id = google_project.this.project_id
  region         = "us-east1"
  zones          = ["us-east1-b", "us-east1-d"]
  cidr           = "10.67.0.0/21"

  load_balancers = {
    private = {
      enabled = true
    }
  }

  node_groups = [
    {
      node_type         = "e2-standard-2"
      capacity_per_zone = 10
      reservations      = ["SYSTEM", "ZOOKEEPER"]
    },
    {
      node_type         = "n2d-standard-2"
      capacity_per_zone = 10
      reservations      = ["CLICKHOUSE"]
    }
  ]

  peering_connections = {
    project_id   = "peering-project-id"  # Replace with actual peering project ID
    network_name = "peering-network-name"  # Replace with actual peering network name
  }
}

// Since the environment provisioning is an async process, this data source is used to wait for environment to be fully provisioned.
data "altinitycloud_env_gcp_status" "this" {
  name                           = altinitycloud_env_gcp.this.name
  wait_for_applied_spec_revision = altinitycloud_env_gcp.this.spec_revision
}
