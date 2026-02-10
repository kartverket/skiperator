#!/bin/bash

CERT_DIR="$1"
CONTEXT="$2"

if [ -z "$CERT_DIR" ] || [ -z "$CONTEXT" ]; then
  echo "Usage: $0 <cert-dir> <kube-context>" >&2
  exit 1
fi

echo "Extracting webhook certificates for local development..."
echo "Using context: $CONTEXT"
echo "Target directory: $CERT_DIR"

SECRET_NAME="skiperator-webhook-server-cert"

echo "Waiting until secret '$SECRET_NAME' is present in cluster $CONTEXT with valid contents"

kubectl wait -n skiperator-system --context "$CONTEXT" "secret/$SECRET_NAME" --for=jsonpath='{.data.tls\.crt}' --timeout=180s || {
  echo "❌ Timed out waiting for tls.crt" >&2
  exit 1
}

kubectl wait -n skiperator-system --context "$CONTEXT" "secret/$SECRET_NAME" --for=jsonpath='{.data.tls\.key}' --timeout=180s || {
  echo "❌ Timed out waiting for tls.key" >&2
  exit 1
}

mkdir -p "$CERT_DIR"

kubectl get secret "$SECRET_NAME" -n skiperator-system --context "$CONTEXT" \
    -o jsonpath='{.data.tls\.crt}' | base64 -d > "$CERT_DIR/tls.crt"

kubectl get secret "$SECRET_NAME" -n skiperator-system --context "$CONTEXT" \
    -o jsonpath='{.data.tls\.key}' | base64 -d > "$CERT_DIR/tls.key"

echo "✅ Certificates extracted to $CERT_DIR"
