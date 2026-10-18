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

## Release Runbook

### Triggering a New Release
Releases are automated via GitHub Actions (`.github/workflows/release.yml`) and trigger whenever a semantic version tag is pushed:

```bash
# 1. Ensure master is clean and up-to-date
git checkout master
git pull origin master

# 2. Tag the release
git tag -a v0.1.0 -m "Release v0.1.0"

# 3. Push the tag to trigger the pipeline
git push origin v0.1.0
```

### Release Pipeline Stages
1. **Test Verification:** Runs project test suite gates prior to container build.
2. **Container Image Build:** Builds the production container image (`ghcr.io/sumitdalavi/k8s-admission-webhook-from-scratch:<tag>`).
3. **Trivy Vulnerability Scan:** Scans the built container for CRITICAL and HIGH CVEs via Aqua Security Trivy.
4. **Anchore Syft SBOM Generation:** Creates an SPDX JSON Software Bill of Materials (`sbom.spdx.json`).
5. **GitHub Release Publishing:** Creates the release with auto-generated release notes and attaches `sbom.spdx.json`.

### Verifying the Release
```bash
# Check workflow run status
gh run list --workflow=release.yml --repo SumitDalavi/k8s-admission-webhook-from-scratch

# Inspect the published release and attached assets
gh release view v0.1.0 --repo SumitDalavi/k8s-admission-webhook-from-scratch

# Download and inspect the SBOM
gh release download v0.1.0 --repo SumitDalavi/k8s-admission-webhook-from-scratch -p "sbom.spdx.json"
cat sbom.spdx.json | jq .spdxVersion
```

### Release Failure Modes
| Symptom | Cause | Resolution |
|---|---|---|
| `Unable to resolve action ... unable to find version` | Pinned action tag missing or upstream tag modified | Pin action to a verified release tag (`@v0.36.0`) or commit SHA |
| `invalid reference format: repository name must be lowercase` | Container repository owner contains uppercase characters | Lowercase the repository owner before tagging (`tr '[:upper:]' '[:lower:]'`) |
| `File not found: sbom.spdx.json` during release step | SBOM action configured with `artifact-name` instead of `output-file` | Ensure `output-file: sbom.spdx.json` is set in Anchore action |
