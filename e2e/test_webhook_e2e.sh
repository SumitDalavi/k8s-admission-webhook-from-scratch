#!/usr/bin/env bash
set -euo pipefail

KIND_CLUSTER="webhook-e2e-test"
log() { echo "[e2e] $*"; }

log "Creating KIND cluster: $KIND_CLUSTER"
kind create cluster --name "$KIND_CLUSTER" --wait 60s

log "Building and loading container image..."
docker build -t admission-webhook:test .
kind load docker-image admission-webhook:test --name "$KIND_CLUSTER"

log "Generating TLS certs..."
bash scripts/generate-certs.sh

log "Creating namespace and TLS secret..."
kubectl create namespace webhook-system 2>/dev/null || true
kubectl create secret tls webhook-tls \
    --cert=certs/tls.crt --key=certs/tls.key \
    -n webhook-system

log "Deploying webhook..."
kubectl apply -f k8s/

log "Waiting for webhook deployment to be ready..."
kubectl wait --for=condition=Available deployment/admission-webhook -n webhook-system --timeout=60s

# Test: a VALID resource should be admitted
log "Testing with compliant pod..."
kubectl run test-allowed --image=nginx:1.25 -n default --restart=Never

# Test: an INVALID resource must be rejected (non-zero exit on no rejection is a fail)
log "Testing with non-compliant registry (should be REJECTED)..."
if kubectl run test-denied --image=untrusted.registry.io/evil:latest -n default --restart=Never 2>&1 | grep -q "denied"; then
  log "✅ Invalid pod correctly denied"
else
  log "❌ Webhook did not deny invalid pod"
  exit 1
fi

log "E2E Test Passed Successfully!"
kind delete cluster --name "$KIND_CLUSTER"
