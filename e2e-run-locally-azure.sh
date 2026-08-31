#!/usr/bin/env bash
#
# Provisions a Gardener Azure shoot with cloud-manager, runs all @azure
# e2e tests against it, and cleans up afterwards.
#
# Prerequisites:
#   - go, kind, kubectl, yq installed
#   - ./tmp/e2e-config.yaml exists with an azure subscription entry
#
# Usage:
#   ./tools/e2e/e2e-run-locally-azure.sh [extra-godog-tags]
#
# Example:
#   ./tools/e2e/e2e-run-locally-azure.sh "@redis"
#
set -euo pipefail
set -a

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECTROOT=$(realpath "$SCRIPT_DIR/../../")
CONFIG_DIR="$PROJECTROOT/tmp"
GHERKIN_REPORT="$CONFIG_DIR/e2e-azure-report.gherkin"
KIND_CLUSTER_NAME="e2e-azure"
PROVIDER="azure"
TAGS="@azure && ~@skip"
KUBECONFIG=""

if [ -n "${1:-}" ]; then
  TAGS="${1} && $TAGS"
fi

E2E_CONFIG_PATH="$CONFIG_DIR/e2e-config.yaml"

# ---------------------------------------------------------------------------
# helpers
# ---------------------------------------------------------------------------

log() {
  echo "$(date '+%Y-%m-%dT%H:%M:%S') $*"
}

e2e_config_check() {
  log "Config check..."
  if [ ! -f "$E2E_CONFIG_PATH" ]; then
    echo "Config file $E2E_CONFIG_PATH does not exist"
    exit 1
  fi

  local gardenKubeconfig
  gardenKubeconfig=$(yq '.gardenKubeconfig' "$E2E_CONFIG_PATH")
  if [ -z "$gardenKubeconfig" ] || [ "$gardenKubeconfig" = "null" ]; then
    echo "Field gardenKubeconfig is not set in $E2E_CONFIG_PATH"
    exit 1
  fi

  if [ ! -f "$gardenKubeconfig" ]; then
    echo "gardenKubeconfig file $gardenKubeconfig does not exist"
    exit 1
  fi

  local api_resources
  api_resources=$(KUBECONFIG="$gardenKubeconfig" kubectl api-resources 2>&1)
  if ! echo "$api_resources" | grep -q "shoots.*core\.gardener\.cloud/v1beta1"; then
    echo "Shoots kind in core.gardener.cloud/v1beta1 group not found in garden cluster"
    exit 1
  fi

  local azure_count
  azure_count=$(yq '[.subscriptions[] | select(.provider == "azure")] | length' "$E2E_CONFIG_PATH")
  if [ "$azure_count" -lt 1 ]; then
    echo "No subscription with provider: azure found in $E2E_CONFIG_PATH"
    exit 1
  fi
}

credentials_download() {
  log "Downloading credentials..."
  if ! go run ./e2e/cmd credentials download; then
    echo "credentials download failed"
    exit 1
  fi
}

create_kind_cluster() {
  log "Creating kind cluster $KIND_CLUSTER_NAME..."
  if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER_NAME}$"; then
    echo "Error: kind cluster $KIND_CLUSTER_NAME already exists. Delete it first: kind delete cluster --name $KIND_CLUSTER_NAME"
    exit 1
  fi
  kind create cluster --name "$KIND_CLUSTER_NAME" --kubeconfig "$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig.yaml"
  KUBECONFIG="$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig.yaml"
  export KUBECONFIG
}

destroy_kind_cluster() {
  log "Destroying kind cluster $KIND_CLUSTER_NAME..."
  kind delete cluster --name "$KIND_CLUSTER_NAME" || true
}

start_sim() {
  log "Starting SIM..."
  go run ./e2e/cmd sim run > "$CONFIG_DIR/e2e-azure-sim.log" 2>&1 &
  SIM_PID=$!
  echo "SIM_PID=$SIM_PID"
  sleep 5
}

start_cm() {
  log "Starting Cloud Manager..."
  local retries=0
  while lsof -iTCP:8081 -sTCP:LISTEN -t >/dev/null 2>&1; do
    if [ $retries -ge 15 ]; then
      echo "Port 8081 still in use after 15s — aborting"
      exit 1
    fi
    log "Waiting for port 8081 to be free..."
    sleep 1
    retries=$((retries + 1))
  done
  FEATURE_FLAG_CONFIG_FILE="${FEATURE_FLAG_CONFIG_FILE:-$PROJECTROOT/pkg/feature/ff_edge.yaml}" \
  LANDSCAPE="${LANDSCAPE:-dev}" \
  go run ./cmd > "$CONFIG_DIR/e2e-azure-cm.log" 2>&1 &
  CM_PID=$!
  echo "CM_PID=$CM_PID"
  sleep 10
}

create_shared_instance() {
  log "Creating shared Azure instance..."
  go run ./e2e/cmd instance create -a "shared-${PROVIDER}" -p "$PROVIDER" -t 30m -w -v
  go run ./e2e/cmd instance show -a "shared-${PROVIDER}"
  log "Adding cloud-manager module..."
  go run ./e2e/cmd instance modules add -m cloud-manager -a "shared-${PROVIDER}" --wait --verbose --timeout 15m
}

run_tests() {
  log "Running e2e tests with tags: $TAGS"
  RUN_E2E_TESTS=yes go test ./e2e/tests \
    -timeout 0 \
    -v \
    -race \
    -godog.tags "$TAGS" \
    -godog.format "pretty,pretty:$GHERKIN_REPORT"
}

print_report() {
  if [ -f "$GHERKIN_REPORT" ]; then
    echo
    echo "=== Gherkin Report ==="
    cat "$GHERKIN_REPORT" | sed 's/\x1b\[[0-9;]*[mGKH]//g'
    echo "======================"
  fi
}

cleanup_skr() {
  log "Cleaning up SKR resources..."
  if [ -f "$KUBECONFIG" ]; then
    go run ./e2e/cmd instance clean --alias "shared-${PROVIDER}" --verbose --wait --force --timeout 30m || true
    go run ./e2e/cmd instance module remove --alias "shared-${PROVIDER}" --module cloud-manager || true
    sleep 10
  fi
}

delete_shared_instance() {
  log "Deleting shared instance..."
  if [ -f "$KUBECONFIG" ]; then
    go run ./e2e/cmd instance delete -a "shared-${PROVIDER}" -t 60m -w -v || true
    sleep 10
    go run ./e2e/cmd instance clean --alias "shared-${PROVIDER}" --verbose --wait --all --force --timeout 1s || true
  fi
}

cleanup() {
  local exit_code=$?
  echo
  log "Cleanup... (exit_code=$exit_code)"
  print_report
  cleanup_skr
  delete_shared_instance
  [ -n "${SIM_PID:-}" ] && kill "$SIM_PID" 2>/dev/null || true
  [ -n "${CM_PID:-}" ]  && kill "$CM_PID"  2>/dev/null || true
  local waited=0
  while lsof -iTCP:8081 -sTCP:LISTEN -t >/dev/null 2>&1 && [ $waited -lt 15 ]; do
    sleep 1; waited=$((waited + 1))
  done
  destroy_kind_cluster
  exit $exit_code
}

trap cleanup EXIT

# ---------------------------------------------------------------------------
# main
# ---------------------------------------------------------------------------

cd "$PROJECTROOT"
mkdir -p "$CONFIG_DIR"

rm -f "$GHERKIN_REPORT"

e2e_config_check
credentials_download
create_kind_cluster
start_sim
start_cm
create_shared_instance
run_tests

log "Done :)"
