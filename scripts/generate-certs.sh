#!/bin/bash
set -e

echo "================================================="
echo "   Generating TLS Certs for Webhook Server"
echo "================================================="

# Requires openssl
mkdir -p certs
cd certs

cat <<EOF > csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = k8s-admission-webhook
DNS.2 = k8s-admission-webhook.default
DNS.3 = k8s-admission-webhook.default.svc
EOF

echo "Generating private key..."
openssl genrsa -out tls.key 2048

echo "Generating Certificate Signing Request (CSR)..."
openssl req -new -key tls.key -subj "/CN=k8s-admission-webhook.default.svc" -config csr.conf -out tls.csr

echo "Generating self-signed certificate (acting as our own CA for this lab)..."
openssl x509 -req -in tls.csr -signkey tls.key -CAcreateserial -out tls.crt -days 365 -extensions v3_req -extfile csr.conf

cd ..

echo "Creating Kubernetes Secret for the TLS certs..."
kubectl create secret tls webhook-server-tls \
    --cert=certs/tls.crt \
    --key=certs/tls.key \
    --dry-run=client -o yaml > k8s/secret.yaml
kubectl apply -f k8s/secret.yaml

echo "Injecting CA Bundle into WebhookConfigurations..."
CA_BUNDLE=$(cat certs/tls.crt | base64 | tr -d '\n')

# Use sed or awk to replace CA_BUNDLE placeholder
sed "s/\${CA_BUNDLE}/${CA_BUNDLE}/g" k8s/validating-webhook-configuration.yaml.tpl > k8s/validating-webhook-configuration.yaml
sed "s/\${CA_BUNDLE}/${CA_BUNDLE}/g" k8s/mutating-webhook-configuration.yaml.tpl > k8s/mutating-webhook-configuration.yaml

echo "✅ Certificates generated and injected successfully."
