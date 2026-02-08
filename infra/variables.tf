// ============== GENERAL ================
variable "subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "Germany West Central"
}

variable "project_name" {
  description = "Project name for naming resources"
  type        = string
  default     = "myapp"
}

// ============== DATABASE ================
variable "postgres_user" {
  description = "PostgreSQL admin username"
  type        = string
  default     = "postgres"
}

variable "postgres_password" {
  description = "PostgreSQL admin password"
  type        = string
  sensitive   = true
}

variable "postgres_db" {
  description = "PostgreSQL database name"
  type        = string
  default     = "orders"
}

variable "postgres_host" {
  description = "PostgreSQL hostname (use LoadBalancer external IP or K8s internal DNS)"
  type        = string
  default     = "postgres.default.svc.cluster.local"
}

variable "postgres_port" {
  description = "PostgreSQL port"
  type        = number
  default     = 5432
}

variable "postgres_sslmode" {
  description = "Postgres SSL mode"
  type = string
  default = "disable"
}

// ============== STORAGE ================
variable "storage_account_name" {
  description = "Azure Storage Account name (must be globally unique, lowercase, 3-24 chars)"
  type        = string
  default     = "myappstorageinvoice2026"
}

// ============== CONTAINER REGISTRY ================
variable "acr_name" {
  description = "Azure Container Registry name (must be globally unique, lowercase)"
  type        = string
  default     = "myappacrinv2026"
}

// ============== AKS ================
variable "aks_cluster_name" {
  description = "AKS cluster name"
  type        = string
  default     = "myapp-aks-order-invoice"
}

variable "aks_node_count" {
  description = "Number of nodes in default node pool"
  type        = number
  default     = 1
}

variable "aks_vm_size" {
  description = "VM size for AKS nodes"
  type        = string
  default     = "Standard_D2pds_v6"
}

// ============== KEY VAULT ================
variable "key_vault_name" {
  description = "Key Vault name (must be globally unique, 3-24 chars)"
  type        = string
  default     = "myapp-kv-inv2026"
}

// ============== CONTAINER APPS ================
variable "aca_environment_name" {
  description = "Container Apps Environment name"
  type        = string
  default     = "myapp-env"
}

variable "invoice_worker_image" {
  description = "Container image for invoice worker (format: image:tag)"
  type        = string
  default     = "invoice-worker:latest"
}

variable "invoice_worker_cpu" {
  description = "CPU cores for invoice worker container"
  type        = number
  default     = 0.25
}

variable "invoice_worker_memory" {
  description = "Memory for invoice worker container"
  type        = string
  default     = "0.5Gi"
}

variable "invoice_worker_min_replicas" {
  description = "Minimum replicas for invoice worker (0 = scales to zero when idle)"
  type        = number
  default     = 0
}

variable "invoice_worker_max_replicas" {
  description = "Maximum replicas for invoice worker"
  type        = number
  default     = 5
}

// ============== TAGS ================
variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    environment = "dev"
    project     = "order-invoice-system"
    managed_by  = "terraform"
  }
}