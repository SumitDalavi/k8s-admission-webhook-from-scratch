# ADR-002: Validation vs Mutation Scope

## Status: Accepted

## Context
Kubernetes Admission Controllers can operate as `MutatingAdmissionWebhook` (modifies the object) or `ValidatingAdmissionWebhook` (accepts or rejects the object).

## Decision
We chose a strict **Validating Webhook** scope for this project.

## Alternatives Considered
| Option | Pros | Cons | Why rejected |
|---|---|---|---|
| Mutating Webhook | Can automatically fix non-compliant deployments (e.g., inject sidecars) | Harder to debug, ordering matters, side-effects can confuse users | We wanted an explicit "fail-closed" security gate |
| Validating Webhook | Simple binary decision, easy to audit, deterministic | Requires users to manually fix their YAML | **Selected** as it enforces shift-left security directly |

## Consequences
- Positive: Security policies are transparently enforced. If a pod lacks a required label, it is rejected with a clear error message.
- Negative: Higher friction for developers as the system won't automatically fix non-compliant objects for them.
- Trade-offs accepted: We prioritize strict auditing and transparency over developer convenience for this specific security controller.
