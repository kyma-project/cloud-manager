#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

sort_azure() {
  local file="$1"
  local tmp
  tmp=$(mktemp)
  jq '
    .actions        |= sort |
    .notActions     |= sort |
    .dataActions    |= sort |
    .notDataActions |= sort
  ' "$file" > "$tmp"
  mv "$tmp" "$file"
  return 0
}

sort_gcp() {
  local file="$1"
  yq -i '.includedPermissions |= sort' "$file"
  return 0
}

sort_aws() {
  local file="$1"
  local tmp
  tmp=$(mktemp)
  jq '
    .Statement[] |= (
      if has("Action") then .Action |= sort else . end
    )
  ' "$file" > "$tmp"
  mv "$tmp" "$file"
  return 0
}

sort_azure "$SCRIPT_DIR/azure_default.json"
sort_azure "$SCRIPT_DIR/azure_peering.json"

sort_gcp "$SCRIPT_DIR/gcp/gcp_default.yaml"
sort_gcp "$SCRIPT_DIR/gcp/gcp_peering.yaml"

sort_aws "$SCRIPT_DIR/aws/policy-CloudManagerAccess.json"
sort_aws "$SCRIPT_DIR/aws/policy-CloudManagerPeeringAccess.json"
