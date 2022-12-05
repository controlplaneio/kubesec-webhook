---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: kubesec-webhook
  labels:
    app: kubesec-webhook
    kind: validator
webhooks:
  - name: pod.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        - UPDATE
        apiGroups:
        - ""
        apiVersions:
        - "v1"
        resources:
        - pods
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
    sideEffects: None
    timeoutSeconds: 15
    admissionReviewVersions: ["v1"]
  - name: deployment.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        - UPDATE
        apiGroups:
        - apps
        apiVersions:
        - "*"
        resources:
        - deployments
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
    sideEffects: None
    timeoutSeconds: 15
    admissionReviewVersions: ["v1"]
  - name: daemonset.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/"
      caBundle: CA_BUNDLE
    rules:
      - operations:
        - CREATE
        - UPDATE
        apiGroups:
        - apps
        apiVersions:
        - "*"
        resources:
        - daemonsets
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
    sideEffects: None
    timeoutSeconds: 15
    admissionReviewVersions: ["v1"]
  - name: statefulset.admission.kubesc.io
    clientConfig:
      service:
        name: kubesec-webhook
        namespace: kubesec
        path: "/"
      caBundle: CA_BUNDLE
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
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        kubesec-validation: enabled
    sideEffects: None
    timeoutSeconds: 15
    admissionReviewVersions: ["v1"]
