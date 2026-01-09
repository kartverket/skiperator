#!/bin/bash

CERT_DIR="$1"
CONTEXT="$2"

echo "Extracting webhook certificates for local development..."
echo "Using context: $CONTEXT"
echo "Target directory: $CERT_DIR"

mkdir -p "$CERT_DIR"

kubectl get secret skiperator-webhook-server-cert -n skiperator-system --context "$CONTEXT" \
    -o jsonpath='{.data.tls\.crt}' | base64 -d > "$CERT_DIR/tls.crt"

kubectl get secret skiperator-webhook-server-cert -n skiperator-system --context "$CONTEXT" \
    -o jsonpath='{.data.tls\.key}' | base64 -d > "$CERT_DIR/tls.key"

echo "✅ Certificates extracted to $CERT_DIR"
