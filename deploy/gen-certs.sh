#!/usr/bin/env bash

# Defaults
NAMESPACE=${1:-"kubesec"}
NAME=${2:-"kubesec-webhook"}
OS="$(uname)"

# Generate cert

subj="/CN=${NAME}.${NAMESPACE}.svc"
addext="subjectAltName=DNS:${NAME}.${NAMESPACE}.svc"

mkdir -p ./certs

echo "$addext" >> ./certs/kubesec.cnf
echo extendedKeyUsage = serverAuth >> ./certs/kubesec.cnf

docker run --rm -v "${PWD}/certs/":/certs/ --user "$(id -u):$(id -g)" alpine/openssl genrsa -out /certs/webhookCA.key 2048
docker run --rm -v "${PWD}/certs/":/certs/ --user "$(id -u):$(id -g)" alpine/openssl req -new -key /certs/webhookCA.key \
  -subj "$subj" \
  -addext "$addext" \
  -out /certs/webhookCA.csr

docker run --rm -v "${PWD}/certs/":/certs/ --user "$(id -u):$(id -g)" alpine/openssl x509 -req \
  -extfile /certs/kubesec.cnf \
  -days 365 \
  -in /certs/webhookCA.csr \
  -signkey /certs/webhookCA.key \
  -out /certs/webhook.crt

# Generate cert secret
kubectl -n kubesec create secret generic \
    "${NAME}"-certs \
    --from-file=key.pem=./certs/webhookCA.key \
    --from-file=cert.pem=./certs/webhook.crt \
    --dry-run=client -o yaml > ./webhook-certs.yaml

# Encode CABundle
if [[ "$OS" == "Darwin" ]]; then
    CA_BUNDLE=$(cat ./certs/webhook.crt | base64)
elif [[ "$OS" == "Linux" ]]; then
    CA_BUNDLE=$(cat ./certs/webhook.crt | base64 -w0)
else
    echo "Unsupported OS ${OS}"
    exit 1
fi

# Generate Kubernetes webhook registration
sed "s/CA_BUNDLE/${CA_BUNDLE}/" ./webhook-registration.yaml.tpl > ./webhook-registration.yaml

# Clean
rm -rf ./certs/
