# Kubernetes Admission Webhook (Hand-Built, Go)

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
├── Dockerfile              # Containerizes the Go server
└── README.md               # This file
```

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

## Author

**Sumit Dalavi — Senior DevSecOps / Platform Engineer**
- [GitHub](https://github.com/your-username)
- [LinkedIn](https://linkedin.com/in/your-profile)


## CI & Reliability Updates (August 2026)

- **CI Pipeline Remediation:** Successfully resolved all CI/CD pipeline failures and established baseline CI workflows.
- **Specific Fix:** Added and configured robust GitHub Actions workflows for automated testing, linting, and formatting.
- **Status:** 🟩 Passing
