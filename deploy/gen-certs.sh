#!/usr/bin/env bash

# Defaults
NAMESPACE=${1:-"kubesec"}
NAME=${2:-"kubesec-webhook"}
OS="`uname`"

# Generate cert
openssl genrsa -out webhookCA.key 2048
openssl req -new -key ./webhookCA.key \
  -subj "/CN=${NAME}.${NAMESPACE}.svc" \
  -addext "subjectAltName = DNS:${NAME}.${NAMESPACE}.svc" \
  -out ./webhookCA.csr
openssl x509 -req -days 365 -in webhookCA.csr -signkey webhookCA.key -out webhook.crt

# Generate cert secret
kubectl -n kubesec create secret generic \
    ${NAME}-certs \
    --from-file=key.pem=./webhookCA.key \
    --from-file=cert.pem=./webhook.crt \
    --dry-run -o yaml > ./webhook-certs.yaml

# Encode CABundle
if [[ "$OS" == "Darwin" ]]; then
    CA_BUNDLE=$(cat ./webhook.crt | base64)
elif [[ "$OS" == "Linux" ]]; then
    CA_BUNDLE=$(cat ./webhook.crt | base64 -w0)
else
    echo "Unsupported OS ${OS}"
    exit 1
fi

# Generate Kubernetes webhook registration
sed "s/CA_BUNDLE/${CA_BUNDLE}/" ./webhook-registration.yaml.tpl > ./webhook-registration.yaml

# Clean
rm ./webhookCA* && rm ./webhook.crt
