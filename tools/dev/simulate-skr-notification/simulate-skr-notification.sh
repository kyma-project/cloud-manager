#!/usr/bin/env bash
#
# Simulate a runtime-watcher notification to the local cloud-manager listener.
#
# The listener (SKREventListener, /v2/cloud-manager/event) does NOT read the
# runtime-id from the JSON body. It reads it from the CommonName (CN) of the
# client certificate carried in the X-Forwarded-Client-Cert (XFCC) header —
# which Istio injects in production. Locally we forge a self-signed cert whose
# CN == the kymaName we want to notify, URL-encode its PEM, and put it in XFCC.
#
# Usage:
#   ./tools/dev/simulate-skr-notification/simulate-skr-notification.sh <kymaName> [addr]
#
#   <kymaName>  the SKR runtime-id (== KCP Kyma CR name) to notify.
#   [addr]      listener base URL. Default: http://localhost:8083
#
# Example:
#   ./tools/dev/simulate-skr-notification/simulate-skr-notification.sh my-kyma-runtime-id
#
# See the README.md in this directory for details.
#
set -euo pipefail

KYMA_NAME="${1:?usage: $0 <kymaName> [addr]}"
ADDR="${2:-http://localhost:8083}"
COMPONENT="cloud-manager"
URL="${ADDR}/v2/${COMPONENT}/event"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# 1) Self-signed cert with Subject CN == kymaName (this is what becomes runtime-id).
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$tmp/key.pem" -out "$tmp/cert.pem" \
  -days 1 -subj "/CN=${KYMA_NAME}" >/dev/null 2>&1

# 2) URL-encode the PEM (XFCC carries the cert URL-escaped inside Cert="...").
encoded_cert="$(python3 -c 'import sys,urllib.parse; print(urllib.parse.quote(open(sys.argv[1]).read(), safe=""))' "$tmp/cert.pem")"
xfcc="Cert=\"${encoded_cert}\""

# 3) Minimal WatchEvent body. cloud-manager ignores the payload beyond identifying
#    the SKR (runtime-id comes from the cert), but the listener still json-unmarshals it.
body='{"watched":{"name":"example","namespace":"kyma-system"},"watchedGvk":{"group":"cloud-resources.kyma-project.io","version":"v1beta1","kind":"IpRange"}}'

echo "POST ${URL}"
echo "  runtime-id (cert CN) = ${KYMA_NAME}"
echo

curl -sS -i -X POST "$URL" \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-Client-Cert: ${xfcc}" \
  --data "$body"
echo
