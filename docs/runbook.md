# Runbook — k8s-admission-webhook-from-scratch
> Last updated: 2026-08-29

## Prerequisites
| Tool | Required Version | How to check |
|---|---|---|
| Go | >= 1.21 | `go version` |
| kubectl | >= 1.28 | `kubectl version --client` |
| OpenSSL | Latest | `openssl version` |

## Quick Start
```bash
# Start a cluster
kind create cluster --name webhook-lab

# Generate certs
./scripts/generate-certs.sh

# Deploy
docker build -t k8s-admission-webhook:local .
kind load docker-image k8s-admission-webhook:local --name webhook-lab
kubectl apply -f k8s/
kubectl label namespace default admission-webhook=enabled
```

## Run Tests
```bash
# Unit tests
go test ./... -v
```

Expected output:
```
?       github.com/SumitDalavi/k8s-admission-webhook-from-scratch/cmd/webhook   [no test files]
ok      github.com/SumitDalavi/k8s-admission-webhook-from-scratch/pkg/mutate    0.012s
ok      github.com/SumitDalavi/k8s-admission-webhook-from-scratch/pkg/validate  0.011s
```

## Environment Variables
| Variable | Default | Purpose |
|---|---|---|
| TLS_CERT_FILE | `/etc/webhook/certs/tls.crt` | Path to TLS cert |
| TLS_KEY_FILE | `/etc/webhook/certs/tls.key` | Path to TLS key |

## Common Failure Modes
| Symptom | Cause | Fix |
|---|---|---|
| `x509: certificate signed by unknown authority` | CA bundle mismatch | Re-run `./scripts/generate-certs.sh` and re-apply configs |
| `connection refused` on apply | Webhook pod not running or `failurePolicy` is Fail | Check `kubectl get pods -n default` |
