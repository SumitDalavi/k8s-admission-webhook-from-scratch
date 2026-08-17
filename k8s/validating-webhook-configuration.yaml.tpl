apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: k8s-admission-webhook-validation
webhooks:
  - name: validate.k8s-admission-webhook.local
    clientConfig:
      service:
        name: k8s-admission-webhook
        namespace: default
        path: "/validate"
      caBundle: ${CA_BUNDLE}
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Fail
    sideEffects: None
    admissionReviewVersions: ["v1"]
    namespaceSelector:
      matchLabels:
        admission-webhook: enabled
