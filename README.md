Slobodan Obradovic E2 37/2025

// dobijemo identity id 
az aks show -g myapp-orders-invoices -n myapp-aks-order-invoice \
  --query identityProfile.kubeletidentity.clientId -o tsv
