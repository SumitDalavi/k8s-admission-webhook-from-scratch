# Kubernetes Admission Webhook (Hand-Built, Go)

[![CI Pipeline](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SumitDalavi/k8s-admission-webhook-from-scratch?color=0969da&label=release&logo=github)](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/releases)
[![Release Workflow](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/actions/workflows/release.yml/badge.svg)](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/actions/workflows/release.yml)
[![GHCR Container](https://img.shields.io/badge/ghcr.io-v0.2.0-blue?logo=docker&logoColor=white)](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/pkgs/container/k8s-admission-webhook-from-scratch)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![SBOM](https://img.shields.io/badge/SBOM-SPDX%202.3-blueviolet?logo=spdx)](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/releases/tag/v0.2.0)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

[![Go Reference](https://pkg.go.dev/badge/github.com/SumitDalavi/k8s-admission-webhook-from-scratch.svg)](https://pkg.go.dev/github.com/SumitDalavi/k8s-admission-webhook-from-scratch)

> **Maturity:** Full Prototype
> _End-to-end Kubernetes validating and mutating admission webhook built in Go._

A hand-written Kubernetes Validating and Mutating Admission Webhook built from scratch in Go. This project demonstrates a deep understanding of the raw Kubernetes admission control API and TLS handling without relying on policy frameworks like Kyverno or OPA Gatekeeper.

## The Problem
Most engineers who claim Kubernetes security experience have only configured existing policy engines (Kyverno, OPA Gatekeeper) via YAML, without understanding how the Kubernetes API server actually calls out to webhooks, what the `AdmissionReview` request/response contract looks like, or how to handle webhook failure modes (`failurePolicy: Fail` vs `Ignore`) safely in production.

## The Solution
This project implements the underlying HTTP contract directly in Go:
1. **Validating Webhook**: Rejects Pods that run as root, are privileged, or lack a required `team` label.
2. **Mutating Webhook**: Automatically injects a default CPU/Memory resource limit into containers if missing, and adds a `cost-center` label.

```text
+-------------------+       +-----------------------+       +-------------------+
|                   |       |                       |       |                   |
|   `kubectl apply` | ----> |   K8s API Server      | ----> |   Mutating/       |
|                   |       |                       |       |   Validating      |
+-------------------+       +-----------------------+       |   Webhook (Go)    |
                                      |                     |                   |
                                      <----(Allow/Deny)---- +-------------------+
```

## Why This Over the Obvious Alternative
Kyverno and OPA Gatekeeper are the right production choice for most teams. However, building the raw webhook once demonstrates a level of Kubernetes internals understanding that policy-framework configuration alone does not — this is explicitly a differentiator for senior platform engineering roles.

## Tech Stack
- **Language**: Go 1.21+
- **Kubernetes Client**: `client-go`, `k8s.io/api/admission/v1`
- **TLS**: OpenSSL script for self-signed certificates acting as the CA
- **Local Cluster**: `kind`

## Decision Log

| Component | Decision | Rationale |
| :--- | :--- | :--- |
| **Failure Policy (Validating)** | `Fail` | Security policies (blocking root) must fail closed. If the webhook is down, Pod creation is halted to prevent security breaches. |
| **Failure Policy (Mutating)** | `Ignore` | Injecting cost labels and default resources is a "nice-to-have" operation. We should not block production deployments if the mutating webhook is temporarily unavailable. |
| **Language** | Go | Go is the native language of Kubernetes. Using `client-go` allows direct use of the `AdmissionReview` and `Pod` structs. |

## Project Structure

```text
k8s-admission-webhook-from-scratch/
├── cmd/webhook/            # Main HTTP server entrypoint
├── pkg/mutate/             # Mutating logic and unit tests
├── pkg/validate/           # Validating logic and unit tests
├── k8s/                    # Webhook deployments and test pods
├── scripts/                # TLS generation script
├── docs/                   # Documentation and ADRs
├── Dockerfile              # Containerizes the Go server
└── README.md               # This file
```

## 🚀 Progressive Credential Loading & Execution Tiers

This project implements **Progressive Credential Loading (PCL)**, allowing it to run immediately out-of-the-box without external dependencies while seamlessly upgrading as credentials are provided.

### Tier 1: Zero-Config Mock Mode (Default)
Ideal for CI, testing, and rapid local evaluation.
- **Execution:** Run with default configuration (`cp .env.example .env`). Zero API keys, databases, or cloud credentials required.
- **Fallback Behavior:** In-memory caches, deterministic offline heuristics, mock vector embeddings, and local mock storage are automatically used.
- **Health Verification:** Endpoints report `{"mode": "mock"}` for unconfigured dependencies.

### Tier 2: Local Development with Live Services
Connect specific external services while keeping others mocked.
- **Execution:** Provide credentials in `.env` for the services you want active (e.g., `database connection, API keys`).
- **Progressive Upgrade:** The runtime inspects environment variables on boot and dynamically promotes only configured modules to live mode.
- **Safe Fallback:** Any missing credentials automatically remain on mock fallbacks with no crashes or missing service errors.

### Tier 3: Production Hardened
Enterprise deployment with production-grade backends, security scans, and container distributions.
- **OCI Container Image:** Pre-built, multi-arch image on GitHub Container Registry:
  ```bash
  docker pull ghcr.io/sumitdalavi/k8s-admission-webhook-from-scratch:v0.2.0
  ```
- **Production Orchestration:** Mount secrets via Kubernetes Secrets or cloud secret managers with atomic config reload support.
- **Release Verification:** Checksums and SPDX SBOM available on [GitHub Releases](https://github.com/SumitDalavi/k8s-admission-webhook-from-scratch/releases).

## 📚 Documentation

- [Architecture](docs/architecture.md) — System diagram and component details
- [Runbook](docs/runbook.md) — Setup, commands, and expected outputs
- [Decisions](docs/decisions.md) — ADRs for webhook pattern choices
- [Changelog](docs/changelog.md) — Change history

## Prerequisites

| Tool | Purpose |
| :--- | :--- |
| Go 1.21+ | Running tests locally |
| Docker & kind | Building the image and running the local cluster |
| kubectl | Applying manifests and viewing pod state |
| OpenSSL | Generating TLS certificates |

## Step-by-Step Setup

1. **Start a local cluster**:
   ```bash
   kind create cluster --name webhook-lab
   ```

2. **Generate TLS Certificates**:
   ```bash
   # Generates certs and dynamically populates the CA Bundle in the manifests
   ./scripts/generate-certs.sh
   ```

3. **Build and Load the Docker Image**:
   ```bash
   docker build -t k8s-admission-webhook:local .
   kind load docker-image k8s-admission-webhook:local --name webhook-lab
   ```

4. **Deploy the Webhook**:
   ```bash
   # Ensure you have run generate-certs.sh first to create the secret.yaml
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   kubectl apply -f k8s/validating-webhook-configuration.yaml
   kubectl apply -f k8s/mutating-webhook-configuration.yaml
   ```

5. **Enable Webhooks on the default namespace**:
   ```bash
   kubectl label namespace default admission-webhook=enabled
   ```

## Usage & Demo

Apply the test pods to see the webhook in action:
```bash
kubectl apply -f k8s/test-pods.yaml
```

**What happens?**
- `valid-pod`: Will be **ACCEPTED** and mutated! Check it with `kubectl get pod valid-pod -o yaml`. You'll see the `cost-center` label injected and the `200m`/`256Mi` resource limits added to the container.
- `invalid-pod-root`: Will be **REJECTED**. The API server will throw an error: `Error from server: ... Container 'my-app' must set securityContext.runAsNonRoot to true`.
- `invalid-pod-no-team`: Will be **REJECTED**. The API server will throw an error: `Error from server: ... Pod is missing the required 'team' label`.

## Verification

| Check | Expected Result |
| :--- | :--- |
| Unit Tests | `go test ./...` passes. |
| Webhook Deployment | `kubectl get pods -l app=k8s-admission-webhook` shows `1/1` running. |
| Mutation | `kubectl get pod valid-pod -o jsonpath='{.spec.containers[0].resources.limits}'` shows the injected limits. |

## Benchmark Results (Last Run: 2026-08-29)
| Metric | Value | Environment |
|---|---|---|
| Malformed Payload Denial | 100% | Adversارڈial test suite |
| Invalid Pod Rejection | Validated | `kind` E2E Test |
| Valid Pod Admission | Validated | `kind` E2E Test |

## Key Design Decisions
- **Why Custom Bash Cert Generation over cert-manager:** This project prioritizes educational transparency. The manual script exposes the exact mechanics of generating the CA bundle and injecting it into the `ValidatingWebhookConfiguration` object.
- See `docs/adr/` for full Architecture Decision Records.
- See `docs/slo.md` for availability and latency objectives.

## Test Coverage
Includes standard unit tests and adversarial checks against expired TLS or garbage payloads.

### 🧪 Integration Testing & CI
This repository is equipped with local integration testing scripts and a GitHub Actions pipeline (`.github/workflows/ci.yml`).

To run the local mock API server script which spins up the webhook and sends requests:
```bash
chmod +x scripts/test_local.sh
./scripts/test_local.sh
```

To run the Go integration tests manually:
```bash
go test -v ./tests/...
```

## Known Limitations & Honest Scope
- **Certificate Rotation**: Because it avoids `cert-manager`, the certificates are static for 1 year. The script must be manually re-run upon expiration. This is suitable for a prototype but unacceptable for a production cluster.

## 🔒 Release & Security Verification

Every tagged release (
*.*.*) triggers an automated DevSecOps release workflow:
1. **Automated Testing:** Runs go test -v ./... across unit and integration suites.
2. **Container Image Build:** Builds the minimal admission webhook container image (ghcr.io/sumitdalavi/k8s-admission-webhook-from-scratch:<tag>).
3. **Vulnerability Scanning:** Scans the webhook container image using [Aqua Security Trivy](https://github.com/aquasecurity/trivy-action) for HIGH and CRITICAL CVEs.
4. **SBOM Generation:** Generates a Software Bill of Materials (SBOM) in SPDX JSON format (sbom.spdx.json) using [Anchore Syft](https://github.com/anchore/sbom-action).
5. **GitHub Release Publishing:** Packages release notes and attaches sbom.spdx.json as a verifiable supply-chain asset.

## Author

**Sumit Dalavi — Senior DevSecOps / Platform Engineer**
- [GitHub](https://github.com/SumitDalavi)
- [LinkedIn](https://in.linkedin.com/in/sumit-dalavi-762838129)

## CI & Reliability Updates (August 2026)

- **CI Pipeline Remediation:** Successfully resolved all CI/CD pipeline failures and established baseline CI workflows.
- **Specific Fix:** Added and configured robust GitHub Actions workflows for automated testing, linting, and formatting.
- **Status:** 🟩 Passing

## Release & Security Verification

Every tagged release (`v*.*.*`) triggers an automated DevSecOps release workflow:
1. **Automated Testing:** Runs test suite gates before build.
2. **Container Image Build:** Builds production container image (`ghcr.io/sumitdalavi/k8s-admission-webhook-from-scratch:<tag>`).
3. **Vulnerability Scanning:** Scans the image using [Aqua Security Trivy](https://github.com/aquasecurity/trivy-action) for `HIGH` and `CRITICAL` CVEs.
4. **SBOM Generation:** Generates a Software Bill of Materials (SBOM) in SPDX JSON format (`sbom.spdx.json`) using [Anchore Syft](https://github.com/anchore/sbom-action).
5. **GitHub Release Publishing:** Packages release notes and attaches `sbom.spdx.json` as a verifiable supply-chain asset.

### Verifying Releases via GitHub CLI
```bash
# View release details
gh release view v0.1.0 --repo SumitDalavi/k8s-admission-webhook-from-scratch

# Download and verify SBOM
gh release download v0.1.0 --repo SumitDalavi/k8s-admission-webhook-from-scratch --pattern "sbom.spdx.json"
cat sbom.spdx.json | jq .
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Runbook](docs/runbook.md)
- [Decisions](docs/decisions.md)
- [Changelog](docs/changelog.md)
