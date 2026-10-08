#!/usr/bin/env bash

# Local e2e runner for AliCloud (SKR NFS).
#
# Adapted from e2e-run-locally-os.sh (OpenStack). Runs the full local stack in tmux:
#   kind cluster (acts as KCP) -> sim -> cloud-manager -> shared SKR instance -> godog tests -> teardown.
#
# Prerequisites:
#   - tmp/e2e-config.yaml with a valid gardenKubeconfig and an `alicloud` subscription
#   - AliCloud credentials for cloud-manager: ALICLOUD_ACCESS_KEY / ALICLOUD_SECRET_KEY
#     (exported below into the .env sourced by the cloud-manager tmux window)
#   - kind, tmux, yq, go on PATH

set -euo pipefail
set -a
#set -o xtrace

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECTROOT=$(realpath "$SCRIPT_DIR/../../")
CONFIG_DIR="$PROJECTROOT/tmp"
GHERKIN_REPORT="$CONFIG_DIR/e2e-alicloud-report.gherkin"
RUN_E2E_TESTS=yes

E2E_CONFIG_PATH="$CONFIG_DIR/e2e-config.yaml"
KUBECONFIG=""
TMUX_SESSION="e2e-alicloud"
KIND_CLUSTER_NAME="e2e-alicloud"
# Node image must be K8s >= 1.31 — the cloud-control CRDs use the CEL isCIDR() function
# which older apiservers reject with "undeclared reference to 'isCIDR'". Repo targets 1.35
# (Makefile ENVTEST_K8S_VERSION). Pin the exact image+digest that ships with kind 0.32.0 —
# a mismatched digest hangs at "Starting control-plane". Override via KIND_NODE_IMAGE if you
# use a different kind version (see that kind release's notes for its node image digests).
KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-kindest/node:v1.33.4@sha256:25a6018e48dfcaee478f4a59af81157a437f15e6e140bf103f85a2e7cd0cbbf2}"

PROVIDER="alicloud"
ALIAS="shared-alicloud"
GODOG_TAGS="@skr && @alicloud && @nfs"

# REUSE=1 keeps the kind cluster + shared instance (Garden shoot) alive across runs:
# it provisions whatever is missing, always runs the tests, and skips teardown on exit
# (only the sim/cloud-manager tmux windows are restarted). The first REUSE run provisions
# everything and keeps it; later REUSE runs reuse it — avoiding the ~10-30m shoot wait.
# A plain run (REUSE unset) does the full provision -> test -> destroy lifecycle.
REUSE="${REUSE:-}"

e2e_config_check() {
  echo "Config check..."
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

  if [ -z "${ALICLOUD_ACCESS_KEY:-}" ] || [ -z "${ALICLOUD_SECRET_KEY:-}" ]; then
    echo "ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY must be set (cloud-manager needs them to call the AliCloud API)"
    exit 1
  fi

  local runtime_config
  runtime_config=$(go run ./e2e/cmd config dump)
  runtime_gardenKubeconfig=$(echo "$runtime_config" | yq '.gardenKubeconfig')
  if [[ "$runtime_gardenKubeconfig" != "$gardenKubeconfig" ]]; then
    echo "runtime didn't read gardenKubeconfig: runtime = $runtime_gardenKubeconfig config = $gardenKubeconfig"
    exit 1
  fi
}

create_dot_env() {
  cp /dev/null "$CONFIG_DIR/.env"
  {
    echo "export CONFIG_DIR=$CONFIG_DIR"
    echo "export FEATURE_FLAG_CONFIG_FILE=$PROJECTROOT/pkg/feature/ff_edge.yaml"
    echo "export ALICLOUD_ACCESS_KEY=$ALICLOUD_ACCESS_KEY"
    echo "export ALICLOUD_SECRET_KEY=$ALICLOUD_SECRET_KEY"
    echo "export SKR_RUNTIME_CONCURRENCY=3"
  } >> "$CONFIG_DIR/.env"
}

create_kind_cluster() {
  if [ -n "$REUSE" ] && kind get clusters | grep -q "$KIND_CLUSTER_NAME"; then
    echo "Reusing existing kind cluster $KIND_CLUSTER_NAME ..."
    KUBECONFIG="$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig"
    kind export kubeconfig --name "$KIND_CLUSTER_NAME" --kubeconfig "$KUBECONFIG"
    return
  fi
  echo "Creating kind cluster $KIND_CLUSTER_NAME (node image $KIND_NODE_IMAGE) ..."
  if kind get clusters | grep "$KIND_CLUSTER_NAME"; then
    echo "Error: Kind cluster $KIND_CLUSTER_NAME already exist. Ensure kind cluster with such name does not exist"
    exit 1
  fi
  kind create cluster --name "$KIND_CLUSTER_NAME" --image "$KIND_NODE_IMAGE" --kubeconfig "$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig"
  KUBECONFIG="$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig"
}

destroy_kind_cluster() {
  echo "Destroying kind cluster $KIND_CLUSTER_NAME ..."
  kind delete cluster --name "$KIND_CLUSTER_NAME"
}

# tmux_run_window <window_name> <command...>
# Creates (or reuses) the tmux session and runs the command in a new named window.
tmux_run_window() {
  local window_name="$1"
  shift
  if tmux list-windows -t "$TMUX_SESSION" 2>&1 | grep "$window_name"; then
    echo "Error: tmux window $window_name already exists"
    exit 1
  fi

  tmux new-session -d -s "$TMUX_SESSION" -x 120 -y 40 -e "KUBECONFIG=$KUBECONFIG" -e "CONFIG_DIR=$CONFIG_DIR" 2>/dev/null || true
  tmux new-window -t "$TMUX_SESSION" -n "$window_name"
  tmux set-window-option -t "$TMUX_SESSION:$window_name" remain-on-exit on
  tmux send-keys -t "$TMUX_SESSION:$window_name" "source $CONFIG_DIR/.env" Enter
  tmux send-keys -t "$TMUX_SESSION:$window_name" "$*; exit \$?" Enter
}

# tmux_teardown
# Kills all windows and the entire tmux session.
tmux_teardown() {
  echo "Killing tmux sessions..."
  tmux kill-session -t "$TMUX_SESSION" 2>/dev/null || true
}

credentials_download() {
  echo "Downloading credentials ..."
  if ! go run ./e2e/cmd credentials download; then
    echo "credentials download has failed"
    exit 1
  fi
}

start_sim() {
  echo "Starting SIM..."
  tmux_run_window "sim" "go run ./e2e/cmd sim run > $CONFIG_DIR/e2e-alicloud-sim.log 2>&1"
}

start_cm() {
  echo "Starting CloudManager..."
  tmux_run_window "cm" "go run ./cmd > $CONFIG_DIR/e2e-alicloud-cm.log 2>&1"
}

create_shared_instance() {
  echo "Creating shared instance..."
  go run ./e2e/cmd instance create -a "$ALIAS" -p "$PROVIDER" -t 30m -w -v
  go run ./e2e/cmd instance show -a "$ALIAS"
  echo "Adding CloudManager module"
  go run ./e2e/cmd instance modules add -m cloud-manager -a "$ALIAS" --wait --verbose --timeout 5m
}

run_tests() {
  echo "Running tests..."
  go test ./e2e/tests -timeout 0 -v -race -godog.tags "$GODOG_TAGS" -godog.format "pretty,pretty:$GHERKIN_REPORT"
}

delete_shared_instance() {
  echo "Cleaning shared instance..."
  go run ./e2e/cmd instance clean --alias "$ALIAS" --verbose --wait --force --timeout 30m
  echo "Removing CloudManager Module..."
  go run ./e2e/cmd instance module remove --alias "$ALIAS" --module cloud-manager
  sleep 10
  echo "Deleting shared instance..."
  go run ./e2e/cmd instance delete -a "$ALIAS" -t 60m -w -v
  sleep 10
  echo "Force-deleting any leftovers..."
  go run ./e2e/cmd instance clean --alias "$ALIAS" --verbose --wait --all --force --timeout 1s || true
}

cleanup() {
  echo
  echo "Cleanup..."
  tmux_teardown
  if [ -n "$REUSE" ]; then
    echo "REUSE set — keeping kind cluster $KIND_CLUSTER_NAME and shared instance alive."
    return
  fi
  destroy_kind_cluster
}

trap cleanup EXIT


e2e_config_check
create_dot_env
credentials_download
create_kind_cluster
start_sim
start_cm

echo "Idling..."
sleep 10

if [ -z "$REUSE" ]; then
  create_shared_instance
else
  # provision the instance only if it is not already present in the reused KCP
  if go run ./e2e/cmd instance show -a "$ALIAS" >/dev/null 2>&1; then
    echo "Reusing existing shared instance $ALIAS ..."
    # cloud-manager was just (re)started — wait until it has connected to the SKR and the
    # cloud-manager module is Ready (which is when the SKR CRDs are installed) before running
    # tests. Without this the test races ahead of CRD install and fails with "no matches for kind".
    # `modules add --wait` is idempotent: it no-ops if already added and just waits for Ready.
    # 10m: a freshly restarted cloud-manager flaps the module Processing->Error->Ready during
    # startup; the wait tolerates Error (keeps polling) but needs enough window to reach Ready.
    echo "Waiting for cloud-manager module to be Ready on the reused SKR ..."
    go run ./e2e/cmd instance modules add -m cloud-manager -a "$ALIAS" --wait --verbose --timeout 10m
  else
    echo "Shared instance $ALIAS not found — provisioning it (first REUSE run) ..."
    create_shared_instance
  fi
fi

run_tests

if [ -z "$REUSE" ]; then
  delete_shared_instance
fi

echo "Done :)"
