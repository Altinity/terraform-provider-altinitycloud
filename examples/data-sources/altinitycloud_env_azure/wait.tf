resource "altinitycloud_env_azure" "azure" {
  name            = "acme-staging"
  cidr            = "10.136.0.0/21"
  region          = "eastus"
  zones           = ["eastus-1", "eastus-2"]
  tenant_id       = "f3c1e3cb-3d92-4315-b98c-0a66676da2e8"
  subscription_id = "3f919947-3102-4210-82ee-4d2ca69f2a01"

  node_groups = [{
    node_type         = "Standard_B2s_v2"
    capacity_per_zone = 3
    reservations      = ["CLICKHOUSE", "ZOOKEEPER", "SYSTEM"]
  }]
}

data "altinitycloud_env_azure_status" "current" {
  name                           = altinitycloud_env_azure.this.name
  wait_for_applied_spec_revision = altinitycloud_env_azure.this.spec_revision
}
