---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: kubesec-webhook
  labels:
    app: kubesec-webhook
    kind: validator
webhooks:
  - name: deployment.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/deployment"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        - UPDATE
        apiGroups:
        - apps
        - extensions
        apiVersions:
        - "*"
        resources:
        - deployments
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
  - name: daemonset.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/daemonset"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        - UPDATE
        apiGroups:
        - apps
        - extensions
        apiVersions:
        - "*"
        resources:
        - daemonsets
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
  - name: statefulset.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/statefulset"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        apiGroups:
        - apps
        apiVersions:
        - "*"
        resources:
        - statefulsets
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
