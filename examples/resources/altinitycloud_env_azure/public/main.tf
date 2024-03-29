provider "azurerm" {
  skip_provider_registration = true
  features {}
}

variable "subscription_id" {
  type        = string
  description = "The ID of subscription to connect to Altinity.Cloud"
}

variable "tenant_id" {
  type        = string
  description = "The ID of tenant to connect to Altinity.Cloud"
}

data "azuread_client_config" "current" {}

data "azuread_service_principal" "altinity_cloud" {
  # Do not change this client_id
  client_id = "8ce5881c-ff0f-47f7-b391-931fbac6cd4b"
}

resource "random_uuid" "azurerm_role_assignment_altinity_cloud" {}

resource "azurerm_role_assignment" "altinity_cloud" {
  name                 = random_uuid.azurerm_role_assignment_altinity_cloud.id
  scope                = "/subscriptions/${var.subscription_id}"
  role_definition_name = "Owner"
  principal_id         = data.azuread_service_principal.altinity_cloud.object_id
}

locals {
  region = "eastus"
  zones  = ["eastus-1", "eastus-2"]
}

resource "altinitycloud_env_azure" "azure" {
  name            = "acme-staging"
  cidr            = "10.136.0.0/21"
  region          = local.region
  tenant_id       = var.tenant_id
  subscription_id = var.subscription_id
  zones           = local.zones

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
