#!/usr/bin/env python3

from azure.storage.queue import QueueServiceClient
from azure.storage.blob import BlobServiceClient, PublicAccess

# Azurite connection string
connection_string = (
    "DefaultEndpointsProtocol=http;"
    "AccountName=devstoreaccount1;"
    "AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;"
    "BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
    "QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1;"
)

print("🔵 Creating Blob container with public access...")
blob_service = BlobServiceClient.from_connection_string(connection_string)
try:
    # Create container with PUBLIC READ access for blobs
    blob_service.create_container("invoices", public_access=PublicAccess.Blob)
    print("✅ Blob container 'invoices' created with public access")
except Exception as e:
    if "ContainerAlreadyExists" in str(e):
        print("⚠️  Container already exists, setting public access...")
        # Set public access on existing container
        container_client = blob_service.get_container_client("invoices")
        container_client.set_container_access_policy(signed_identifiers={}, public_access=PublicAccess.Blob)
        print("✅ Public access enabled on 'invoices' container")
    else:
        print(f"❌ Error: {e}")

print("🔵 Creating Queue...")
queue_service = QueueServiceClient.from_connection_string(connection_string)
try:
    queue_service.create_queue("invoice-queue")
    print("✅ Queue 'invoice-queue' created")
except Exception as e:
    if "QueueAlreadyExists" in str(e):
        print("✅ Queue 'invoice-queue' already exists")
    else:
        print(f"❌ Error: {e}")

print("✅ Azure storage initialization complete!")