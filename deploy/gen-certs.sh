#!/usr/bin/env bash

# Defaults
WEBHOOK_NS=${1:-"default"}
NAME=${2:-"kubesec"}
WEBHOOK_SVC="${NAME}-webhook"
OS="`uname`"

# Generate certs
openssl genrsa -out webhookCA.key 2048
openssl req -new -key ./webhookCA.key -subj "/CN=${WEBHOOK_SVC}.${WEBHOOK_NS}.svc" -out ./webhookCA.csr
openssl x509 -req -days 365 -in webhookCA.csr -signkey webhookCA.key -out webhook.crt

# Generate Kubernetes secret
kubectl create secret generic \
    ${WEBHOOK_SVC}-certs \
    --from-file=key.pem=./webhookCA.key \
    --from-file=cert.pem=./webhook.crt \
    --dry-run -o yaml > ./webhook-certs.yaml

# Set the CABundle on the webhook registration
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
