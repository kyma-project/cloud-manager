#!/usr/bin/env bash
#
# Provisions a Gardener AliCloud shoot with cloud-manager, runs all @alicloud
# e2e tests against it, and cleans up afterwards.
#
# Prerequisites:
#   - go, kind, kubectl, yq installed
#   - ./tmp/e2e-config.yaml exists with an alicloud subscription entry
#     (see e2e-config-example.yaml)
#   - ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY exported, or present in
#     ./tmp/ALICLOUD_ACCESS_KEY and ./tmp/ALICLOUD_SECRET_KEY files
#
# Usage:
#   ./e2e-run-locally-alicloud.sh [extra-godog-tags]
#
# Example:
#   ./e2e-run-locally-alicloud.sh "@redis"
#
set -euo pipefail
set -a

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECTROOT=$(realpath "$SCRIPT_DIR")
CONFIG_DIR="$PROJECTROOT/tmp"
GHERKIN_REPORT="$CONFIG_DIR/e2e-alicloud-report.gherkin"
KIND_CLUSTER_NAME="e2e-alicloud"
PROVIDER="alicloud"
TAGS="@alicloud && ~@skip"
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

  local alicloud_count
  alicloud_count=$(yq '[.subscriptions[] | select(.provider == "alicloud")] | length' "$E2E_CONFIG_PATH")
  if [ "$alicloud_count" -lt 1 ]; then
    echo "No subscription with provider: alicloud found in $E2E_CONFIG_PATH"
    exit 1
  fi
}

check_alicloud_credentials() {
  log "Checking AliCloud credentials..."

  # Accept from env or from files dropped by credentials download
  if [ -z "${ALICLOUD_ACCESS_KEY:-}" ] && [ -f "$CONFIG_DIR/ALICLOUD_ACCESS_KEY" ]; then
    ALICLOUD_ACCESS_KEY=$(cat "$CONFIG_DIR/ALICLOUD_ACCESS_KEY")
  fi
  if [ -z "${ALICLOUD_SECRET_KEY:-}" ] && [ -f "$CONFIG_DIR/ALICLOUD_SECRET_KEY" ]; then
    ALICLOUD_SECRET_KEY=$(cat "$CONFIG_DIR/ALICLOUD_SECRET_KEY")
  fi

  if [ -z "${ALICLOUD_ACCESS_KEY:-}" ] || [ -z "${ALICLOUD_SECRET_KEY:-}" ]; then
    echo "ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY must be set (env or ./tmp/ALICLOUD_ACCESS_KEY|ALICLOUD_SECRET_KEY files)"
    exit 1
  fi

  export ALICLOUD_ACCESS_KEY
  export ALICLOUD_SECRET_KEY
}

credentials_download() {
  log "Downloading credentials..."
  if ! go run ./e2e/cmd credentials download; then
    echo "credentials download failed"
    exit 1
  fi
  # Use the downloaded cloud-manager service account key for the CM process,
  # not the personal ALICLOUD_ACCESS_KEY used for infrastructure operations.
  if [ -f "$CONFIG_DIR/ALICLOUD_ACCESS_KEY" ]; then
    CM_ALICLOUD_ACCESS_KEY=$(cat "$CONFIG_DIR/ALICLOUD_ACCESS_KEY")
    CM_ALICLOUD_SECRET_KEY=$(cat "$CONFIG_DIR/ALICLOUD_SECRET_KEY")
    export CM_ALICLOUD_ACCESS_KEY CM_ALICLOUD_SECRET_KEY
    log "CM will use cloud-manager service account credentials from downloaded secret"
  fi
}

create_kind_cluster() {
  log "Creating kind cluster $KIND_CLUSTER_NAME..."
  if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER_NAME}$"; then
    log "Kind cluster $KIND_CLUSTER_NAME already exists — reusing it."
  else
    kind create cluster --name "$KIND_CLUSTER_NAME" --kubeconfig "$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig.yaml"
  fi
  KUBECONFIG="$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig.yaml"
  export KUBECONFIG
}

destroy_kind_cluster() {
  log "Destroying kind cluster $KIND_CLUSTER_NAME..."
  kind delete cluster --name "$KIND_CLUSTER_NAME" || true
}

stop_sim() {
  log "Stopping any running SIM process..."
  local pids
  pids=$(pgrep -f "go run ./e2e/cmd sim run" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
    sleep 2
  fi
  [ -n "${SIM_PID:-}" ] && kill "$SIM_PID" 2>/dev/null || true
  SIM_PID=""
}

stop_cm() {
  log "Stopping any running Cloud Manager process..."
  # Kill the go run wrapper and any compiled 'cmd' child holding port 8081.
  local pids
  pids=$(pgrep -f "go run ./cmd" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
  fi
  [ -n "${CM_PID:-}" ] && kill "$CM_PID" 2>/dev/null || true
  CM_PID=""
  # Also kill whatever process is listening on 8081 (the compiled binary child).
  local waited=0
  while lsof -iTCP:8081 -sTCP:LISTEN -t >/dev/null 2>&1 && [ $waited -lt 30 ]; do
    local port_pids
    port_pids=$(lsof -iTCP:8081 -sTCP:LISTEN -t 2>/dev/null || true)
    [ -n "$port_pids" ] && kill $port_pids 2>/dev/null || true
    sleep 1; waited=$((waited + 1))
  done
  if lsof -iTCP:8081 -sTCP:LISTEN -t >/dev/null 2>&1; then
    local port_pids
    port_pids=$(lsof -iTCP:8081 -sTCP:LISTEN -t 2>/dev/null || true)
    [ -n "$port_pids" ] && kill -9 $port_pids 2>/dev/null || true
    sleep 2
  fi
}

start_sim() {
  stop_sim
  log "Starting SIM..."
  go run ./e2e/cmd sim run > "$CONFIG_DIR/e2e-alicloud-sim.log" 2>&1 &
  SIM_PID=$!
  echo "SIM_PID=$SIM_PID"
  sleep 5
}

start_cm() {
  stop_cm
  log "Starting Cloud Manager..."
  # Wait for health-probe port to be free (leftover from a previous run).
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
  ALICLOUD_ACCESS_KEY="${CM_ALICLOUD_ACCESS_KEY:-$ALICLOUD_ACCESS_KEY}" \
  ALICLOUD_SECRET_KEY="${CM_ALICLOUD_SECRET_KEY:-$ALICLOUD_SECRET_KEY}" \
  go run ./cmd > "$CONFIG_DIR/e2e-alicloud-cm.log" 2>&1 &
  CM_PID=$!
  echo "CM_PID=$CM_PID"
  sleep 10
}

create_shared_instance() {
  log "Creating shared AliCloud instance..."
  go run ./e2e/cmd instance create -a "shared-${PROVIDER}" -p "$PROVIDER" -t 30m -s 5m -w -v
  go run ./e2e/cmd instance show -a "shared-${PROVIDER}"
  log "Waiting 30s for SKR API server to stabilise before adding module..."
  sleep 30
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

cleanup_kcp_finalizers() {
  log "Force-removing finalizers from any stuck KCP redis objects..."
  if [ ! -f "$KUBECONFIG" ]; then return; fi
  for kind in redisinstance rediscluster; do
    local names
    names=$(kubectl --kubeconfig="$KUBECONFIG" get "$kind" -n kcp-system \
      -o jsonpath='{range .items[?(@.metadata.deletionTimestamp)]}{.metadata.name}{"\n"}{end}' 2>/dev/null || true)
    for name in $names; do
      log "Removing finalizers from stuck $kind/$name"
      kubectl --kubeconfig="$KUBECONFIG" patch "$kind" "$name" -n kcp-system \
        --type=json -p='[{"op":"remove","path":"/metadata/finalizers"}]' 2>/dev/null || true
    done
  done
}

cleanup() {
  local exit_code=$?
  echo
  log "Cleanup... (exit_code=$exit_code)"
  print_report
  cleanup_skr
  cleanup_kcp_finalizers
  delete_shared_instance
  stop_sim
  stop_cm
  destroy_kind_cluster
  exit $exit_code
}

trap cleanup EXIT

# ---------------------------------------------------------------------------
# main
# ---------------------------------------------------------------------------

cd "$PROJECTROOT"
mkdir -p "$CONFIG_DIR"

# Remove stale report from a previous run so cleanup never prints ghost results.
rm -f "$GHERKIN_REPORT"

e2e_config_check
credentials_download
check_alicloud_credentials
create_kind_cluster
start_sim
start_cm
create_shared_instance
run_tests

log "Done :)"
