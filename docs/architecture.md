# Architecture — k8s-admission-webhook-from-scratch
> Last updated: 2026-08-29 | Maturity: Full Prototype
> _Kubernetes Validating and Mutating Admission Webhook in Go._

## System Diagram
The following Mermaid.js sequence diagram maps the core workflow and interactions:

```mermaid
flowchart TD
    Dev(["Developer"])
    API["kube-apiserver"]
    Webhook["Mutating/Validating Webhook (Go)"]

    Dev -->|"kubectl apply Pod"| API
    API -->|"AdmissionReview (Mutating)"| Webhook
    Webhook -->|"Patch: Inject limits & labels"| API
    API -->|"AdmissionReview (Validating)"| Webhook
    Webhook -->|"Allow / Deny (Security Context)"| API
    API -->|"Create / Reject Pod"| Dev
```

## Component Table

| Component | File | Responsibility | Tech |
|---|---|---|---|
| Main Server | `cmd/webhook/main.go` | HTTP server listening for AdmissionReview requests | Go |
| Mutator | `pkg/mutate/mutate.go` | Injects labels and default resource limits | Go |
| Validator | `pkg/validate/validate.go` | Rejects privileged, root pods, or missing labels | Go |

## Port Assignments

| Service | Port | Notes |
|---|---|---|
| Webhook Server | `8443` | TLS enabled endpoint for AdmissionReview |

## Dependency Honesty Table

| Dependency | Status | Notes |
|---|---|---|
| Kubernetes API Server | **Real** | Webhook integrates directly with the API server. |
| OpenSSL | **Real** | Used to generate self-signed certs for local dev. |
| kind (Local Cluster) | **Optional** | Used for local testing and demo scripts. |

## Progressive Credential Loading & Production Architecture (v0.2.0)

This service implements the **Progressive Credential Loading (PCL)** pattern, decoupling development speed from external infrastructure provisioning:

```mermaid
graph LR
    Env[Environment Credentials] -->|Detect Keys & URLs| Detector[Credential Detector]
    Detector -->|No Credentials / Missing| Mock[Mock Mode Factory]
    Detector -->|Valid Credentials Present| Live[Live Provider Factory]
    Mock --> Runtime[Application Core & /health]
    Live --> Runtime
```

### Core Architecture Components
1. **Mode Detection Singleton:** Inspects environment configuration at boot time without hard crashes on missing secrets.
2. **Dual-Mode Factory Wrappers:**
   - **Storage / Cache:** In-process maps or SQLite `:memory:` transparently promote to managed Redis or PostgreSQL when connection strings are supplied.
   - **AI Providers:** Deterministic heuristic generators and local embeddings transparently promote to live OpenAI, Anthropic, or Gemini API clients.
3. **Hot-Reload & Dynamic TLS:** File watchers monitor mounted volume paths (Kubernetes ConfigMaps and Secrets) to update certificates and limit configs atomically without process restarts.
4. **Health & Observability:** The `/health` endpoint exposes real-time runtime mode metadata (`mock` vs `live`) across all backing adapters.
