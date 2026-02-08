Slobodan Obradovic E2 37/2025

// dobijemo identity id 
az aks show -g myapp-orders-invoices -n myapp-aks-order-invoice \
  --query identityProfile.kubeletidentity.clientId -o tsv

  

kubectl get svc postgres // dobijemo postgres ip

// postavimo postgres ip u ACA
POSTGRES_IP=$(kubectl get svc postgres -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo $POSTGRES_IP  # Proveri

# Update ACA environment variable:
az containerapp update \
  --name invoice-worker \
  --resource-group myapp-orders-invoices \
  --set-env-vars DB_HOST=$POSTGRES_IP