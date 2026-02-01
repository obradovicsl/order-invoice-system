terraform {
    required_providers {
        azurerm = {
            source  = "hashicorp/azurerm"
            version = ">= 3.0.0"
        }
    }
}

provider "azurerm" {
    features {}
    subscription_id = "84f19c42-fe0d-4412-90f0-aaf3d6fced80"
}

# Create a resource group
resource "azurerm_resource_group" "main" {
    name = "myapp-orders-invoices"
    location = "Germany West Central"
}

# Create a storage account
resource "azurerm_storage_account" "storage" {
    name = "myappstorageinvoivce2026"
    resource_group_name = azurerm_resource_group.main.name
    location = azurerm_resource_group.main.location
    account_tier = "Standard"
    account_replication_type = "LRS"
}

# Create ACR 
resource "azurerm_container_registry" "acr" {
    name = "myappacrinv2026"
    resource_group_name = azurerm_resource_group.main.name
    location = azurerm_resource_group.main.location
    sku = "Basic"
    admin_enabled = true
}

# Create AKS cluster
resource "azurerm_kubernetes_cluster" "aks" {
    name = "myapp-aks-order-invoice"
    location = azurerm_resource_group.main.location
    resource_group_name = azurerm_resource_group.main.name
    dns_prefix = "myapp"

    default_node_pool {
        name = "default"
        node_count = 1
        vm_size = "Standard_D2pds_v6"
    }

    identity {
        type = "SystemAssigned"
    }

    network_profile {
        network_plugin = "kubenet"
        load_balancer_sku = "standard"
    }
}

# Attach ACR to AKS
resource "azurerm_role_assignment" "aks_acr" {
    principal_id                     = azurerm_kubernetes_cluster.aks.kubelet_identity[0].object_id
    role_definition_name             = "AcrPull"
    scope                            = azurerm_container_registry.acr.id
    skip_service_principal_aad_check = true
}


# Create storage queue
resource "azurerm_storage_queue" "invoice_queue" {
    name = "invoice-queue"
    storage_account_id = azurerm_storage_account.storage.id
}

# Create storage container - Blob
resource "azurerm_storage_container" "invoices" {
    name = "invoices"
    container_access_type = "private"
    storage_account_id = azurerm_storage_account.storage.id
}

# Container Apps Environment
resource "azurerm_container_app_environment" "env" {
    name                = "myapp-env"
    location            = azurerm_resource_group.main.location
    resource_group_name = azurerm_resource_group.main.name
}

output "acr_login_server" {
    value = azurerm_container_registry.acr.login_server
}

output "acr_admin_username" {
    value = azurerm_container_registry.acr.admin_username
}

output "acr_admin_password" {
    value     = azurerm_container_registry.acr.admin_password
    sensitive = true
}

output "storage_connection_string" {
    value     = azurerm_storage_account.storage.primary_connection_string
    sensitive = true
}

output "storage_account_name" {
    value = azurerm_storage_account.storage.name
}

output "storage_account_key" {
    value     = azurerm_storage_account.storage.primary_access_key
    sensitive = true
}

output "aks_cluster_name" {
    value = azurerm_kubernetes_cluster.aks.name
}

output "resource_group_name" {
    value = azurerm_resource_group.main.name
}