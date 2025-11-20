#!/bin/bash
# generate-tls.sh
# Usage: ./generate-tls.sh <release-name> <namespace>

set -e

if [ $# -ne 2 ]; then
  echo "Usage: $0 <release-name> <namespace>"
  exit 1
fi

RELEASE_NAME="$1"
NAMESPACE="$2"
SERVICE_NAME="${RELEASE_NAME}-${RELEASE_NAME}"

KEY_FILE="tls.key"
CRT_FILE="tls.crt"

echo "Generating self-signed TLS cert for release '$RELEASE_NAME' in namespace '$NAMESPACE'..."

openssl genrsa -out "$KEY_FILE" 2048

openssl req -x509 -new -nodes -key "$KEY_FILE" -days 365 \
  -out "$CRT_FILE" \
  -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" \
  -addext "subjectAltName = DNS:${SERVICE_NAME},DNS:${SERVICE_NAME}.${NAMESPACE},DNS:${SERVICE_NAME}.${NAMESPACE}.svc,DNS:${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local"

TLS_KEY_B64=$(base64 -w0 < "$KEY_FILE")
TLS_CRT_B64=$(base64 -w0 < "$CRT_FILE")

echo
echo "Copy these values into your values.yaml:"
echo
echo "tls:"
echo "  key: $TLS_KEY_B64"
echo "  crt: $TLS_CRT_B64"
echo
echo "Done!"
