# Cloud-Native Order & Invoice Platform

A cloud-native microservices application for order management and asynchronous PDF invoice generation, deployed on Azure using AKS, Azure Container Apps, and supporting Azure services.

### Services

| Service | Runtime | Description |
|---------|---------|-------------|
| **Frontend** | React | Dashboard — product catalog, order creation, order tracking with PDF download |
| **Order Service** | Go | Order management, stock validation, queue message publishing |
| **Catalog Service** | Go | Product catalog, inventory tracking |
| **Invoice Worker** | Go | Async PDF generation, Blob Storage upload, order status update |
| **PostgreSQL** | StatefulSet + PVC | Persistent relational storage |

---

## Infrastructure (Terraform)

All Azure infrastructure is provisioned with Terraform.

### Resources provisioned

- **Resource Group**
- **Azure Container Registry (ACR)** — stores all Docker images
- **AKS Cluster** — runs Frontend, Order Service, Catalog Service, PostgreSQL
- **Azure Container Apps Environment + Container App** — runs Invoice Worker
- **Azure Storage Account**
  - Queue Storage — async messaging between Order Service and Invoice Worker
  - Blob Storage — stores generated PDF invoices and product images
- **Azure Key Vault** — centralized secrets (DB credentials, storage connection strings, etc.)
- **AKS Secrets Store CSI Driver** — mounts Key Vault secrets into AKS pods

### Usage

```bash
cd terraform/

# Initialize
terraform init

# Preview changes
terraform plan

# Apply
terraform apply
```

> After `terraform apply`, all resource names and connection strings are output and automatically referenced by the GitHub Actions deployment pipeline.

---

## Getting Started (Local)

### Prerequisites

- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Terraform](https://developer.hashicorp.com/terraform/install)
- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
- Azure subscription

### Local development with Azurite

For local testing without a real Azure account, [Azurite](https://github.com/Azure/Azurite) emulates Azure Blob, Queue, and Table Storage:

```bash
docker run -p 10000:10000 -p 10001:10001 -p 10002:10002 mcr.microsoft.com/azure-storage/azurite
```

Set the following environment variable in your services:

```
AZURE_STORAGE_CONNECTION_STRING=UseDevelopmentStorage=true
```

---

## CI/CD Pipeline (GitHub Actions)

The pipeline runs automatically on every push to `main` and consists of 4 phases:

### Phase 1 — Tests
- Installs dependencies for each microservice
- Runs unit tests (2 per service)
- Build verification
- **Pipeline does not proceed if tests fail**

### Phase 2 — Build
- Builds Docker images for all microservices
- Pushes images to Azure Container Registry (ACR)

### Phase 3 — Deploy
**AKS:**
- Updates Kubernetes Deployment resources with new image tags
- Monitors rollout status (`kubectl rollout status`)
- Verifies pods are running and ready

**Container Apps:**
- Updates Invoice Worker to new image revision
- Verifies new revision is active

### Phase 4 — Post-deploy verification
- Runs 2 integration tests against the live environment
- **Automatic rollback** on AKS if integration tests fail (`kubectl rollout undo`)

---

## Kubernetes Resources (AKS)

| Resource | Type | Notes |
|----------|------|-------|
| `frontend` | Deployment | React app |
| `order-service` | Deployment | Go, exposed via Ingress |
| `catalog-service` | Deployment | Go, exposed via Ingress |
| `postgres` | StatefulSet | Persistent volume claim |
| `ingress` | Ingress | Routes `/api/orders`, `/api/catalog`, `/` |
| `secretproviderclass` | SecretProviderClass | Mounts Key Vault secrets via CSI Driver |

---

## Async Invoice Generation

When an order is created, the Order Service immediately returns a confirmation and publishes a message to **Azure Queue Storage**. The Invoice Worker (running on Container Apps) picks up the message, generates a PDF, uploads it to **Blob Storage**, and updates the order status to `completed`.

This decouples order creation from invoice generation:
- Order Service continues to accept new orders even if the Invoice Worker is temporarily down
- Container Apps scales the worker to **zero replicas** when the queue is empty (no cost), and automatically scales up when messages arrive
- Failed PDF generations are retried automatically (message stays in queue)

---
