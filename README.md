# kubesec-webhook

[![Build Status](https://travis-ci.org/controlplaneio/kubesec-webhook.svg?branch=master)](https://travis-ci.org/controlplaneio/kubesec-webhook)

Kubesec.io admission controller for Kubernetes Deployments, DaemonSets and StatefulSets

For the kubectl scan plugin see [kubectl-kubesec](https://github.com/controlplaneio/kubectl-kubesec)

### Install

Generate webhook configuration files with a new TLS certificate and CA Bundle:

```bash
make certs
```

Deploy the admission controller and webhooks in the kubesec namespace (requires Kubernetes 1.10 or newer):

```bash
make deploy
``` 

Enable Kubesec validation by adding this label:

```bash
kubectl label namespaces default kubesec-validation=enabled
```

### Usage

Try to apply a privileged Deployment:

```bash
kubectl apply -f ./test/deployment.yaml

Error from server (InternalError): error when creating "./test/deployment.yaml": 
Internal error occurred: admission webhook "deployment.admission.kubesec.io" denied the request: 
deployment-test score is -30, deployment minimum accepted score is 0
Scan Result:
{
  "error": "",
  "score": -30,
  "scoring": {
    "critical": [
      {
        "selector": "containers[] .securityContext .privileged == true",
        "reason": "Privileged containers can allow almost completely unrestricted host access",
        "weight": 0
      }
    ],
    "advise": [
      {
        "selector": "containers[] .securityContext .runAsNonRoot == true",
        "reason": "Force the running image to run as a non-root user to ensure least privilege"
      },
      {
        "selector": "containers[] .securityContext .capabilities .drop",
        "reason": "Reducing kernel capabilities available to a container limits its attack surface",
        "href": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
      },
      {
        "selector": "containers[] .securityContext .readOnlyRootFilesystem == true",
        "reason": "An immutable root filesystem can prevent malicious binaries being added to 
PATH and increase attack cost"
      },
      {
        "selector": "containers[] .securityContext .runAsUser \u003e 10000",
        "reason": "Run as a high-UID user to avoid conflicts with the host's user table"
      },
      {
        "selector": "containers[] .securityContext .capabilities .drop | index(\"ALL\")",
        "reason": "Drop all capabilities and add only those required to reduce syscall attack surface"
      }
    ]
  }
}
```

Try to apply a privileged DaemonSet:

```bash
kubectl apply -f ./test/daemonset.yaml

Error from server (InternalError): error when creating "./test/daemonset.yaml": 
Internal error occurred: admission webhook "daemonset.admission.kubesec.io" denied the request: 
daemonset-test score is -30, daemonset minimum accepted score is 0
Scan Result:
{
  "error": "",
  "score": -30,
  "scoring": {
    "critical": [
      {
        "selector": "containers[] .securityContext .privileged == true",
        "reason": "Privileged containers can allow almost completely unrestricted host access",
        "weight": 0
      }
    ],
    "advise": [
      {
        "selector": "containers[] .securityContext .runAsNonRoot == true",
        "reason": "Force the running image to run as a non-root user to ensure least privilege"
      },
      {
        "selector": "containers[] .securityContext .capabilities .drop",
        "reason": "Reducing kernel capabilities available to a container limits its attack surface",
        "href": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
      },
      {
        "selector": "containers[] .securityContext .readOnlyRootFilesystem == true",
        "reason": "An immutable root filesystem can prevent malicious binaries being added to PATH and increase attack cost"
      },
      {
        "selector": "containers[] .securityContext .runAsUser \u003e 10000",
        "reason": "Run as a high-UID user to avoid conflicts with the host's user table"
      },
      {
        "selector": "containers[] .securityContext .capabilities .drop | index(\"ALL\")",
        "reason": "Drop all capabilities and add only those required to reduce syscall attack surface"
      }
    ]
  }
}
```

Try to apply a privileged StatefulSet:

```bash
kubectl apply -f ./test/statefulset.yaml

Error from server (InternalError): error when creating "./test/statefulset.yaml": 
Internal error occurred: admission webhook "statefulset.admission.kubesec.io" denied the request: 
statefulset-test score is -30, statefulset minimum accepted score is 0
Scan Result:
{
  "error": "",
  "score": -30,
  "scoring": {
    "critical": [
      {
        "selector": "containers[] .securityContext .privileged == true",
        "reason": "Privileged containers can allow almost completely unrestricted host access",
        "weight": 0
      }
    ],
    "advise": [
      {
        "selector": ".spec .volumeClaimTemplates[] .spec .accessModes | index(\"ReadWriteOnce\")",
        "reason": ""
      },
      {
        "selector": "containers[] .securityContext .runAsNonRoot == true",
        "reason": "Force the running image to run as a non-root user to ensure least privilege"
      },
      {
        "selector": "containers[] .securityContext .capabilities .drop",
        "reason": "Reducing kernel capabilities available to a container limits its attack surface",
        "href": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
      },
      {
        "selector": "containers[] .securityContext .readOnlyRootFilesystem == true",
        "reason": "An immutable root filesystem can prevent malicious binaries being added to 
PATH and increase attack cost"
      },
      {
        "selector": "containers[] .securityContext .runAsUser \u003e 10000",
        "reason": "Run as a high-UID user to avoid conflicts with the host's user table"
      }
    ]
  }
}
```

### Configuration

You can set the minimum Kubesec.io score in `./deploy/webhook/yaml`:

```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: kubesec-webhook
  labels:
    app: kubesec-webhook
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: kubesec-webhook
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8081"
    spec:
      containers:
        - name: kubesec-webhook
          image: controlplaneio/kubesec:0.1-dev
          imagePullPolicy: Always
          command:
            - ./kubesec
          args:
            - -tls-cert-file=/etc/webhook/certs/cert.pem
            - -tls-key-file=/etc/webhook/certs/key.pem
            - -min-score=0
          ports:
            - containerPort: 8080
            - containerPort: 8081
          volumeMounts:
            - name: webhook-certs
              mountPath: /etc/webhook/certs
              readOnly: true
      volumes:
        - name: webhook-certs
          secret:
            secretName: kubesec-webhook-certs
```

### Monitoring 

The admission controller exposes Prometheus RED metrics for each webhook a Grafana dashboard is available [here](https://grafana.com/dashboards/7088).

### Credits

Kudos to [Xabier](https://github.com/slok) for the awesome [kubewebhook library](https://github.com/slok/kubewebhook).  
