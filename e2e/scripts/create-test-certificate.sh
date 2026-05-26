#!/usr/bin/env bash

# Script to generate a self-signed test certificate for AwsCertificate e2e tests
# This certificate is used by the skr-shared-awscertificate.feature test
#
# Usage:
#   ./e2e/scripts/create-test-certificate.sh
#
# This script generates certificate files and prints the base64-encoded values
# that can be copied into the feature file.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/../tmp/certs"

# Configuration
CERT_DAYS=7300  # 20 years validity
CERT_SUBJECT="/C=US/ST=Test/L=Test/O=E2E Testing/OU=Cloud Manager/CN=e2e-test.example.com"

echo "Creating certificate directory: ${CERT_DIR}"
mkdir -p "${CERT_DIR}"

echo "Generating RSA private key..."
openssl genrsa -out "${CERT_DIR}/tls.key" 2048 2>/dev/null

echo "Generating self-signed certificate..."
openssl req -new -x509 \
  -key "${CERT_DIR}/tls.key" \
  -out "${CERT_DIR}/tls.crt" \
  -days ${CERT_DAYS} \
  -subj "${CERT_SUBJECT}" \
  -addext "subjectAltName=DNS:e2e-test.example.com,DNS:*.e2e-test.example.com" 2>/dev/null

echo "Copying certificate as CA certificate..."
cp "${CERT_DIR}/tls.crt" "${CERT_DIR}/ca.crt"

echo ""
echo "✓ Certificate generated successfully!"
echo ""
echo "Certificate details:"
openssl x509 -in "${CERT_DIR}/tls.crt" -noout -subject -dates

echo ""
echo "Certificate files created at:"
echo "  - ${CERT_DIR}/tls.crt"
echo "  - ${CERT_DIR}/tls.key"
echo "  - ${CERT_DIR}/ca.crt"
echo ""
echo "To create the Kubernetes Secret, run:"
echo "  ./e2e/scripts/create-certificate-secret.sh"
echo ""
