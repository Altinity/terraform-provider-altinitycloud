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
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
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
}
