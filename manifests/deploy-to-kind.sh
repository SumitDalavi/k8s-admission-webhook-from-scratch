#!/usr/bin/env bash
# Local KIND cluster demo for the admission webhook
set -euo pipefail

CLUSTER="${KIND_CLUSTER:-webhook-demo}"
log() { echo "[kind-demo] $*"; }

# 1. Create cluster
log "Creating KIND cluster: $CLUSTER"
kind create cluster --name "$CLUSTER" 2>/dev/null || log "Cluster already exists"

# 2. Build the webhook binary
log "Building webhook..."
go build -o webhook-server ./cmd/webhook/

# 3. Build and load image into KIND
log "Building and loading container image..."
docker build -t admission-webhook:dev .
kind load docker-image admission-webhook:dev --name "$CLUSTER"

# 4. Generate TLS certificates
log "Generating TLS certs..."
bash scripts/generate-certs.sh

# 5. Create namespace and TLS secret
kubectl create namespace webhook-system 2>/dev/null || true
kubectl create secret tls webhook-tls     --cert=certs/tls.crt --key=certs/tls.key     -n webhook-system 2>/dev/null || kubectl delete secret webhook-tls -n webhook-system &&     kubectl create secret tls webhook-tls --cert=certs/tls.crt --key=certs/tls.key -n webhook-system

# 6. Deploy webhook
kubectl apply -f k8s/

# 7. Wait for deployment
kubectl rollout status deployment/admission-webhook -n webhook-system --timeout=60s

# 8. Test: deploy a compliant pod
log "Testing with compliant pod..."
kubectl run test-allowed --image=nginx:1.25 -n default --restart=Never 2>/dev/null || true

# 9. Test: deploy a non-compliant pod (should be rejected by imagepolicy)
log "Testing with non-compliant registry (should be REJECTED)..."
kubectl run test-denied --image=untrusted.registry.io/evil:latest -n default --restart=Never 2>/dev/null &&     log "ERROR: Pod should have been rejected!" || log "PASS: Pod correctly rejected"

log "Demo complete. Cluster: $CLUSTER"
log "Clean up: kind delete cluster --name $CLUSTER"
