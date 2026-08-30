# ADR-001: Certificate Management Approach

## Status: Accepted

## Context
Kubernetes Admission Webhooks require a TLS certificate to securely communicate with the Kubernetes API Server. We must decide how to generate, rotate, and manage this certificate bundle.

## Decision
We chose a **custom bash-based generation approach** integrated into the deployment pipeline, injecting the CA bundle dynamically using a patching script.

## Alternatives Considered
| Option | Pros | Cons | Why rejected |
|---|---|---|---|
| Cert-Manager | Industry standard, fully automated | Heavy dependency to install in a cluster for a single webhook | Overkill for a "from scratch" educational/minimal prototype |
| Kube-Builder / Operator-SDK | Built-in scaffold for certs | Abstracts away the actual mechanics | Conflicts with the "from scratch" philosophy |
| Custom Bash Script + Patching | Zero external dependencies, exposes internal workings clearly | Requires manual script execution during deployment | **Selected** as it perfectly aligns with the repository's educational and lightweight goals |

## Consequences
- Positive: Developers explicitly learn how Kubernetes `ValidatingWebhookConfiguration` caBundle injection works.
- Negative: Certificates are not automatically rotated. If the cert expires after 1 year, the webhook will break until the script is run again.
- Trade-offs accepted: We accept manual rotation in exchange for zero dependencies.
