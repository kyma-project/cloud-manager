#!/usr/bin/env bash

# Script to create Kubernetes Secret from test certificate files
# This Secret is used by the skr-shared-awscertificate.feature test
#
# Prerequisites:
#   - Certificate files must exist in e2e/tmp/certs/ (run create-test-certificate.sh first)
#   - KUBECONFIG must be set to point to the SKR cluster
#
# Usage:
#   ./e2e/scripts/create-certificate-secret.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/../tmp/certs"

SECRET_NAME="e2e-test-certificate"
NAMESPACE="default"

# Verify certificate files exist
if [ ! -f "${CERT_DIR}/tls.crt" ]; then
  echo "Error: Certificate file not found: ${CERT_DIR}/tls.crt"
  echo "Run ./e2e/scripts/create-test-certificate.sh first"
  exit 1
fi

if [ ! -f "${CERT_DIR}/tls.key" ]; then
  echo "Error: Private key file not found: ${CERT_DIR}/tls.key"
  echo "Run ./e2e/scripts/create-test-certificate.sh first"
  exit 1
fi

if [ ! -f "${CERT_DIR}/ca.crt" ]; then
  echo "Error: CA certificate file not found: ${CERT_DIR}/ca.crt"
  echo "Run ./e2e/scripts/create-test-certificate.sh first"
  exit 1
fi

echo "Creating Kubernetes Secret '${SECRET_NAME}' in namespace '${NAMESPACE}'..."

# Create TLS Secret from certificate files
kubectl create secret tls "${SECRET_NAME}" \
  --cert="${CERT_DIR}/tls.crt" \
  --key="${CERT_DIR}/tls.key" \
  --namespace="${NAMESPACE}"

echo "Adding CA certificate to Secret..."

# Detect OS for base64 command
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  CA_CERT_BASE64=$(base64 < "${CERT_DIR}/ca.crt" | tr -d '\n')
else
  # Linux
  CA_CERT_BASE64=$(base64 -w0 < "${CERT_DIR}/ca.crt")
fi

# Add CA certificate to the Secret
kubectl patch secret "${SECRET_NAME}" \
  --namespace="${NAMESPACE}" \
  --type=merge \
  -p "{\"data\":{\"ca.crt\":\"${CA_CERT_BASE64}\"}}"

echo ""
echo "✓ Secret '${SECRET_NAME}' created successfully in namespace '${NAMESPACE}'"
echo ""
echo "Secret contents:"
kubectl get secret "${SECRET_NAME}" -n "${NAMESPACE}" -o jsonpath='{.data}' | jq 'keys'
echo ""
