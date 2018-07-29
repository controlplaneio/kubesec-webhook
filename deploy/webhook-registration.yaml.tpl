---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: kubesec-pod-webhook
  labels:
    app: kubesec-webhook
    kind: validator
webhooks:
  - name: webhook.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: default
        path: "/webhooks/validating/pod"
      caBundle: CA_BUNDLE
    rules:
      - operations: [ "CREATE", "UPDATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: kubesec-deployment-webhook
  labels:
    app: kubesec-webhook
    kind: validator
webhooks:
  - name: webhook.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: default
        path: "/webhooks/validating/deployment"
      caBundle: CA_BUNDLE
    rules:
      - operations: [ "CREATE", "UPDATE" ]
        apiGroups: ["extensions"]
        apiVersions: ["v1beta1"]
        resources: ["deployments"]
      - operations: [ "CREATE", "UPDATE" ]
        apiGroups: ["apps"]
        apiVersions: ["v1beta1", "v1beta2", "v1"]
        resources: ["deployments"]
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
