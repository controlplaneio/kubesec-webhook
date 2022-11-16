---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: kubesec-webhook
  labels:
    app: kubesec-webhook
    kind: validator
webhooks:
- admissionReviewVersions:
  # NOTE(ludo): admission v1 requires github.com/slok/kubewebhook/v2
  #- v1
  - v1beta1
  clientConfig:
    service:
      name: kubesec-webhook
      namespace: kubesec
      path: "/deployment"
    caBundle: CA_BUNDLE
  failurePolicy: Fail
  name: deployment.admission.kubesec.io
  namespaceSelector:
    matchLabels:
      kubesec-validation: enabled
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
  sideEffects: None
- admissionReviewVersions:
  # NOTE(ludo): admission v1 requires github.com/slok/kubewebhook/v2
  #- v1
  - v1beta1
  clientConfig:
    service:
      name: kubesec-webhook
      namespace: kubesec
      path: "/daemonset"
    caBundle: CA_BUNDLE
  failurePolicy: Fail
  namespaceSelector:
    matchLabels:
      kubesec-validation: enabled
  name: daemonset.admission.kubesec.io
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
  sideEffects: None
- admissionReviewVersions:
  # NOTE(ludo): admission v1 requires github.com/slok/kubewebhook/v2
  #- v1
  - v1beta1
  clientConfig:
    service:
      name: kubesec-webhook
      namespace: kubesec
      path: "/statefulset"
    caBundle: CA_BUNDLE
  failurePolicy: Fail
  namespaceSelector:
    matchLabels:
      kubesec-validation: enabled
  name: statefulset.admission.kubesec.io
  rules:
    - operations:
      - CREATE
      - UPDATE
      apiGroups:
      - apps
      apiVersions:
      - "*"
      resources:
      - statefulsets
  sideEffects: None
