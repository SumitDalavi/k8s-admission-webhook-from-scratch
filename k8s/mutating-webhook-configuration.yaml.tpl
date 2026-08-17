apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: k8s-admission-webhook-mutation
webhooks:
  - name: mutate.k8s-admission-webhook.local
    clientConfig:
      service:
        name: k8s-admission-webhook
        namespace: default
        path: "/mutate"
      caBundle: ${CA_BUNDLE}
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Ignore
    sideEffects: None
    admissionReviewVersions: ["v1"]
    namespaceSelector:
      matchLabels:
        admission-webhook: enabled
