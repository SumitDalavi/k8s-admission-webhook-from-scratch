# Service Level Objectives

## Latency SLO
- Target: P99 < 200ms
- Context: The Kubernetes API server halts the admission of a resource while waiting for this webhook to return a decision. A slow webhook degrades the entire cluster's control plane responsiveness.
- Measurement: HTTP request duration on the `/validate` endpoint.

## Availability SLO
- Target: 99.95% uptime
- Error budget: 21.9 minutes/month
- Note: `failurePolicy: Fail` is configured. If this webhook goes down, NO pods that match the namespace selector can be scheduled. Availability is absolutely critical.

## Correctness SLO
- Zero bypasses of invalid payloads.
- Measurement: Audit logs from kube-apiserver cross-referenced against the webhook's internal rejection logs.
