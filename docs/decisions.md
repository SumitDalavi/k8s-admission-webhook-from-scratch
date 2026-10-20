# Decisions

## ADR-001: Failure Policy Configuration
**Date:** 2026-08-29  
**Status:** Accepted

**Context:**  
Admission webhooks must specify a `failurePolicy` (Fail or Ignore) in case the webhook server is unreachable or errors out.

**Decision:**  
The Validating webhook uses `Fail` (fail closed) to ensure security constraints (no root, no privileged) are strictly enforced even if the webhook goes down. The Mutating webhook uses `Ignore` (fail open) because injecting default limits is an operational convenience, and a webhook outage should not block all deployments.

**Consequences:**  
- ✅ Security guarantees are robustly maintained.
- ✅ Resiliency for standard deployments is preserved.
- ⚠️ If the validating webhook goes down, the cluster cannot admit new pods until it is recovered.

## ADR-002: Progressive Credential Loading (PCL) Architecture
- **Context:** Production hardening required live backends (PostgreSQL, Redis, OpenAI, Anthropic), but removing mocks would break CI pipelines and create high developer onboarding friction.
- **Decision:** Implement Progressive Credential Loading. Keep in-memory and local mock heuristics as the default safe baseline, while wrapping external integrations in auto-promoting factory adapters triggered by environment variables.
- **Consequences:** 
  - *Pros:* 100% CI pass rate without requiring external secrets; zero setup friction for new contributors; progressive upgrade to production clusters.
  - *Cons:* Adds lightweight factory abstraction layers and runtime credential detection logic.
