# k8s-admission-webhook-from-scratch Architecture

## System Diagram
The following Mermaid.js sequence diagram maps the core workflow and interactions:

```mermaid
sequenceDiagram
    API_Server->>Webhook: AdmissionReview (POST)
Webhook->>Webhook: Validate specs
Webhook-->>API_Server: AdmissionResponse (Allow/Deny)
```

## Component Breakdown
- **Core Technology**: Go, K8s API
- **Design Paradigm**: Emphasizes high availability, fault tolerance, and security.

## Security & Scaling Considerations
- Strict boundary validations.
- Horizontal scalability achieved via stateless workers.
- Encrypted data at rest and in transit.
