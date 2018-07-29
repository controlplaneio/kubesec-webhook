# kubesec-webhook

Kubesec.io Kubernetes validating admission webhook

### Install

Generate a TLS certificate and CA Bundle with:

```bash
make certs
```

Deploy Kubesec.io admission controller in the default namespace:

```bash
kubectl apply -f ./deploy/
``` 

Enable Kubesec validation by adding this label:

```bash
kubectl label namespaces default kubesec-validation=enabled
```

Try to apply a privileged pod:

```bash
kubectl apply -f ./test/pod.yaml

Error from server (InternalError): error when creating "./test/pod.yaml": 
Internal error occurred: admission webhook "webhook.kubesc.io" denied the request: 
pod-test score is -26, pod minimum accepted score is 0
``` 

Try to apply a privileged deployment:

```bash
kubectl apply -f ./test/deployment.yaml

Error from server (InternalError): error when creating "./test/deployment.yaml": 
Internal error occurred: admission webhook "webhook.kubesc.io" denied the request: 
deployment-test score is -26, pod minimum accepted score is 0
```

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
          image: stefanprodan/kubesec:0.1-test0
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
