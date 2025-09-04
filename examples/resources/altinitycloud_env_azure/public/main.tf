provider "azurerm" {
  skip_provider_registration = true
  features {}
}

locals {
  # Replace these values with your own Azure tenant and subscription IDs
  tenant_id       = "f3c1e3cb-3d92-4315-b98c-0a66676da2e8"
  subscription_id = "3f919947-3102-4210-82ee-4d2ca69f2a01"
}

data "azuread_client_config" "current" {}

data "azuread_service_principal" "altinity_cloud" {
  # Do not change this client_id
  client_id = "8ce5881c-ff0f-47f7-b391-931fbac6cd4b"
}

resource "random_uuid" "azurerm_role_assignment_altinity_cloud" {}

resource "azurerm_role_assignment" "altinity_cloud" {
  name                 = random_uuid.azurerm_role_assignment_altinity_cloud.id
  scope                = "/subscriptions/${local.subscription_id}"
  role_definition_name = "Owner"
  principal_id         = data.azuread_service_principal.altinity_cloud.object_id
}

resource "altinitycloud_env_azure" "azure" {
  name            = "acme-staging"
  cidr            = "10.136.0.0/21"
  region          = "eastus"
  zones           = ["eastus-1", "eastus-2"]
  tenant_id       = local.tenant_id
  subscription_id = local.subscription_id

  load_balancers = {
    public = {
      enabled          = true
      source_ip_ranges = ["0.0.0.0/0"]
    }
  }

  node_groups = [{
    node_type         = "Standard_B2s_v2"
    capacity_per_zone = 3
    reservations      = ["CLICKHOUSE", "ZOOKEEPER", "SYSTEM"]
  }]
}

// ⚠️ Environment provisioning is asynchronous.
// Without this data source, Terraform cannot detect provisioning failures.
// This data source waits until the environment is fully reconciled and reports errors.
data "altinitycloud_env_azure_status" "this" {
  name                           = altinitycloud_env_azure.this.name
  wait_for_applied_spec_revision = altinitycloud_env_azure.this.spec_revision
}
