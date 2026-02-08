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
    subscription_id = var.subscription_id
}

data "azurerm_client_config" "current" {}

// ================ RESOURCE GROUP ================
resource "azurerm_resource_group" "main" {
    name     = "${var.project_name}-orders-invoices"
    location = var.location
    tags     = var.tags
}


// =============== STORAGE ACCOUNT ================
resource "azurerm_storage_account" "storage" {
    name                     = var.storage_account_name
    resource_group_name      = azurerm_resource_group.main.name
    location                 = azurerm_resource_group.main.location
    account_tier             = "Standard"
    account_replication_type = "LRS"
    tags                     = var.tags
}

// Queue
resource "azurerm_storage_queue" "invoice_queue" {
    name = "invoice-queue"
    storage_account_id = azurerm_storage_account.storage.id
}

// Blob
resource "azurerm_storage_container" "invoices" {
    name = "invoices"
    container_access_type = "private"
    storage_account_id = azurerm_storage_account.storage.id
}


// =============== CONTAINER REGISTRY ================
resource "azurerm_container_registry" "acr" {
    name                = var.acr_name
    resource_group_name = azurerm_resource_group.main.name
    location            = azurerm_resource_group.main.location
    sku                 = "Basic"
    admin_enabled       = true
    tags                = var.tags
}

// =============== AKS VIRTUAL NETWORK ================ 
resource "azurerm_virtual_network" "aks_vnet" {
  name = "aks-vnet"
  location = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  address_space = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "aks_subnet" {
  name = "aks-subnet"
  resource_group_name = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.aks_vnet.name
  address_prefixes = ["10.0.1.0/24"]
}

// =============== ACA VIRTUAL NETWORK ================
resource "azurerm_virtual_network" "aca_vnet" {
  name = "aca-vnet"
  location = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  address_space = ["10.1.0.0/16"]
}

resource "azurerm_subnet" "aca_subnet" {
  name = "aca-subnet"
  resource_group_name = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.aca_vnet.name
  address_prefixes = ["10.1.0.0/23"]

  delegation {
    name = "aca-delegation"

    service_delegation {
      name = "Microsoft.App/environments"

      actions = [
        "Microsoft.Network/virtualNetworks/subnets/action"
      ]
    }
  }
}

// =============== PEERING ================

resource "azurerm_virtual_network_peering" "aks_to_aca" {
  name                      = "aks-to-aca"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.aks_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.aca_vnet.id
  allow_forwarded_traffic   = true
  allow_gateway_transit     = false
  allow_virtual_network_access = true
}

resource "azurerm_virtual_network_peering" "aca_to_aks" {
  name                      = "aca-to-aks"
  resource_group_name       = azurerm_resource_group.main.name
  virtual_network_name      = azurerm_virtual_network.aca_vnet.name
  remote_virtual_network_id = azurerm_virtual_network.aks_vnet.id
  allow_forwarded_traffic   = true
  allow_gateway_transit     = false
  allow_virtual_network_access = true
}


// ============== AKS CLUSTER ================
resource "azurerm_kubernetes_cluster" "aks" {
    name                = var.aks_cluster_name
    location            = azurerm_resource_group.main.location
    resource_group_name = azurerm_resource_group.main.name
    dns_prefix          = var.project_name

    default_node_pool {
        name       = "default"
        node_count = var.aks_node_count
        vm_size    = var.aks_vm_size
        vnet_subnet_id = azurerm_subnet.aks_subnet.id
    }

    identity {
        type = "SystemAssigned"
    }

    network_profile {
        network_plugin   = "azure"
        service_cidr = "10.2.0.0/16"
        dns_service_ip = "10.2.0.10"
        load_balancer_sku = "standard"
    }

    key_vault_secrets_provider {
        secret_rotation_enabled  = true
        secret_rotation_interval = "2m"
    }

    tags = var.tags
}

// Add permission for AKS to pull images from ACR
resource "azurerm_role_assignment" "aks_acr" {
    principal_id                     = azurerm_kubernetes_cluster.aks.kubelet_identity[0].object_id
    role_definition_name             = "AcrPull"
    scope                            = azurerm_container_registry.acr.id
    skip_service_principal_aad_check = true
}


// ============== KEY VAULT ================
resource "azurerm_key_vault" "main" {
  name                       = var.key_vault_name
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  purge_protection_enabled   = false
  tags                       = var.tags
}

// Give Terraform user access to create/read secrets
resource "azurerm_key_vault_access_policy" "terraform" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  secret_permissions = [
    "Get",
    "List",
    "Set",
    "Delete",
    "Purge"
  ]
}

// Postgres password
resource "azurerm_key_vault_secret" "postgres_password" {
  name         = "postgres-password"
  value        = var.postgres_password
  key_vault_id = azurerm_key_vault.main.id
}

// Postgres user
resource "azurerm_key_vault_secret" "postgres_user" {
  name         = "postgres-user"
  value        = var.postgres_user
  key_vault_id = azurerm_key_vault.main.id
}

// Postgres db_name
resource "azurerm_key_vault_secret" "postgres_db" {
  name         = "postgres-db"
  value        = var.postgres_db
  key_vault_id = azurerm_key_vault.main.id
}

// Storage connection string
resource "azurerm_key_vault_secret" "storage_connection_string" {
  name         = "storage-connection-string"
  value        = azurerm_storage_account.storage.primary_connection_string
  key_vault_id = azurerm_key_vault.main.id
}

// Allow AKS to read secrets from Key Vault
resource "azurerm_key_vault_access_policy" "aks" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_kubernetes_cluster.aks.kubelet_identity[0].object_id

  secret_permissions = [
    "Get",
    "List"
  ]
}

// ============== CONTAINER APPS ENVIRONMENT ================
resource "azurerm_container_app_environment" "env" {
    name                = var.aca_environment_name
    location            = azurerm_resource_group.main.location
    resource_group_name = azurerm_resource_group.main.name
    infrastructure_subnet_id = azurerm_subnet.aca_subnet.id
    internal_load_balancer_enabled = false
    tags                = var.tags
}

// Create a user assigned identity
resource "azurerm_user_assigned_identity" "aca_identity" {
  name                = "${var.project_name}-aca-identity"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  tags                = var.tags
}

// Add access policy for ACA identity to read secrets from Key Vault
resource "azurerm_key_vault_access_policy" "aca" {
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_user_assigned_identity.aca_identity.principal_id

  secret_permissions = [
    "Get",
    "List"
  ]
}

resource "azurerm_role_assignment" "aca_acr_pull" {
    principal_id = azurerm_user_assigned_identity.aca_identity.principal_id
    role_definition_name = "AcrPull"
    scope = azurerm_container_registry.acr.id   
}

// ============== CONTAINER APPS ================
resource "azurerm_container_app" "app" {
    name                         = "invoice-worker"
    container_app_environment_id = azurerm_container_app_environment.env.id
    resource_group_name          = azurerm_resource_group.main.name
    revision_mode                = "Single" 

    identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aca_identity.id]
    }

    registry {
    server   = azurerm_container_registry.acr.login_server
    identity = azurerm_user_assigned_identity.aca_identity.id
    }

    secret {
    name                = "postgres-user"
    key_vault_secret_id = azurerm_key_vault_secret.postgres_user.id
    identity            = azurerm_user_assigned_identity.aca_identity.id
    }

  secret {
    name                = "postgres-password"
    key_vault_secret_id = azurerm_key_vault_secret.postgres_password.id
    identity            = azurerm_user_assigned_identity.aca_identity.id
  }

  secret {
    name                = "postgres-db"
    key_vault_secret_id = azurerm_key_vault_secret.postgres_db.id
    identity            = azurerm_user_assigned_identity.aca_identity.id
  }

  secret {
    name                = "storage-connection-string"
    key_vault_secret_id = azurerm_key_vault_secret.storage_connection_string.id
    identity            = azurerm_user_assigned_identity.aca_identity.id
  }

  template {
   container {
      name   = "invoice-worker"
      image  = "${azurerm_container_registry.acr.login_server}/${var.invoice_worker_image}"
      cpu    = var.invoice_worker_cpu
      memory = var.invoice_worker_memory

      env {
        name        = "DB_USER"
        secret_name = "postgres-user"
      }
      env {
        name        = "DB_PASSWORD"
        secret_name = "postgres-password"
      }
      env {
        name        = "DB_NAME"
        secret_name = "postgres-db"
      }
      env {
        name        = "AZURE_STORAGE_CONNECTION_STRING"
        secret_name = "storage-connection-string"
      }

      env {
        name  = "DB_HOST"
        value = var.postgres_host
      }
      env {
        name  = "DB_PORT"
        value = tostring(var.postgres_port)
      }
      env {
        name  = "DB_SSLMODE"
        value = tostring(var.postgres_sslmode)
      }
      env {
        name  = "BLOB_CONTAINER_NAME"
        value = azurerm_storage_container.invoices.name
      }

      env {
        name  = "QUEUE_NAME"
        value = azurerm_storage_queue.invoice_queue.name
      }
      env {
        name  = "ENV"
        value = var.environment
      }
      env {
        name  = "LOG_LEVEL"
        value = "debug"
      }
   }

    min_replicas = var.invoice_worker_min_replicas
    max_replicas = var.invoice_worker_max_replicas
  }
}

// ============== OUTPUTS ================
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