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
