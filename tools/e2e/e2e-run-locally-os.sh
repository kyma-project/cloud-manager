#!/usr/bin/env bash

set -euo pipefail
set -a
#set -o xtrace

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECTROOT=$(realpath "$SCRIPT_DIR/../../")
CONFIG_DIR="$PROJECTROOT/tmp"
GHERKIN_REPORT="$CONFIG_DIR/e2e-os-report.gherkin"
RUN_E2E_TESTS=yes

E2E_CONFIG_PATH="$CONFIG_DIR/e2e-config.yaml"
KUBECONFIG=""
TMUX_SESSION="e2e-os"
KIND_CLUSTER_NAME="e2e-os"

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

  local openstack_count
  openstack_count=$(yq '[.subscriptions[] | select(.provider == "openstack")] | length' "$E2E_CONFIG_PATH")
  if [ "$openstack_count" -lt 1 ]; then
    echo "No subscription with provider: openstack found in $E2E_CONFIG_PATH"
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
    echo "export GCP_SA_JSON_KEY_PATH=$CONFIG_DIR/gcp-default.json"
    echo "export GCP_VPC_PEERING_KEY_PATH=$CONFIG_DIR/gcp-peering.json"
    echo "export AWS_ROLE_NAME=CloudManagerRole"
    echo "export AWS_PEERING_ROLE_NAME=CloudManagerPeeringRole"
    echo "export SKR_RUNTIME_CONCURRENCY=3"
  } >> "$CONFIG_DIR/.env"
}

create_kind_cluster() {
  echo "Creating kind cluster $KIND_CLUSTER_NAME ..."
  if kind get clusters | grep "$KIND_CLUSTER_NAME"; then
    echo "Error: Kind cluster $KIND_CLUSTER_NAME already exist. Ensure kind cluster with such name does not exist"
    exit 1
  fi
  kind create cluster --name "$KIND_CLUSTER_NAME" --kubeconfig "$CONFIG_DIR/$KIND_CLUSTER_NAME-kubeconfig"
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
  if tmux list-windows -t e2e-os 2>&1 | grep "$window_name"; then
    echo "Error: tmux window $window_name already exists"
    exit 1
  fi

  tmux new-session -d -s "$TMUX_SESSION" -x 120 -y 40 -e "KUBECONFIG=$KUBECONFIG" -e "CONFIG_DIR=$CONFIG_DIR" 2>/dev/null || true
  tmux new-window -t "$TMUX_SESSION" -n "$window_name"
  tmux set-window-option -t "$TMUX_SESSION:$window_name" remain-on-exit on
  tmux send-keys -t "$TMUX_SESSION:$window_name" "source $CONFIG_DIR/.env" Enter
  tmux send-keys -t "$TMUX_SESSION:$window_name" "$*; exit \$?" Enter
}

# tmux_capture_window <window_name> <log_file>
# Captures the full scrollback of a tmux window and stores it in the named variable.
tmux_capture_window() {
  local window_name="$1"
  local log_file="$2"
  local output
  output=$(tmux capture-pane -t "$TMUX_SESSION:$window_name" -p -J -S -)
  echo "$output" > "$log_file"
}

# tmux_window_status <window_name>
# Prints "running" if the window's process is still alive, or "finished <exit_code>" if it has exited.
tmux_window_status() {
  local window_name="$1"
  local dead exit_code
  dead=$(tmux display-message -t "$TMUX_SESSION:$window_name" -p '#{pane_dead}')
  if [ "$dead" = "0" ]; then
    echo "running"
  else
    exit_code=$(tmux display-message -t "$TMUX_SESSION:$window_name" -p '#{pane_dead_status}')
    echo "finished $exit_code"
  fi
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

check_connectivity() {
  local gardener_kuberconfig_path
  gardener_kuberconfig_path=$(cat "$E2E_CONFIG_PATH" | yq '.gardenKubeconfig')
  KUBECONFIG=$gardener_kuberconfig_path kubectl get cloudprofile converged-cloud -o yaml > "$CONFIG_DIR/cloudprofile-converged-cloud.yaml"
  local os_auth_url
  os_auth_url=$(cat "$CONFIG_DIR/cloudprofile-converged-cloud.yaml" | yq '.spec.providerConfig.keystoneURLs[] | select(.region == "eu-de-1") | .url ')
  if ! curl "$os_auth_url/auth/tokens"; then
    echo "Failed to connect to $os_auth_url. Check connectivity! Maybe some VPN issue?"
    exit 1
  fi
}

start_sim() {
  echo "Starting SIM..."
  tmux_run_window "sim" go run ./e2e/cmd sim run | tee "$CONFIG_DIR/e2e-os-sim.log"
}

start_cm() {
  echo "Starting CloudManager..."
  tmux_run_window "cm" go run ./cmd | tee "$CONFIG_DIR/e2e-os-cm.log"
}

create_shared_instance() {
  echo "Creating shared instance..."
  go run ./e2e/cmd instance create -a shared-openstack -p openstack -t 30m -w -v
  go run ./e2e/cmd instance show -a shared-openstack
  echo "Adding CloudManager module"
  go run ./e2e/cmd instance modules add -m cloud-manager -a shared-openstack --wait --verbose --timeout 5m
}

run_tests() {
  echo "Running tests..."
  go test ./e2e/tests -timeout 0 -v -race -godog.tags "@openstack" -godog.format "pretty,pretty:$GHERKIN_REPORT"
}

delete_shared_instance() {
  echo "Cleaning shared instance..."
  go run ./e2e/cmd instance clean --alias shared-openstack --verbose --wait --force --timeout 30m
  echo "Removing CloudManager Module..."
  go run ./e2e/cmd instance module remove --alias shared-openstack --module cloud-manager
  sleep 10
  echo "Deleting shared instance..."
  go run ./e2e/cmd instance delete -a shared-openstack -t 60m -w -v
  sleep 10
  echo "Force-deleting any leftovers..."
  go run ./e2e/cmd instance clean --alias shared-openstack --verbose --wait --all --force --timeout 1s || true
}

cleanup() {
  echo
  echo "Cleanup..."
  tmux_teardown
  destroy_kind_cluster
}

trap cleanup EXIT


e2e_config_check
create_dot_env
credentials_download
check_connectivity
create_kind_cluster
start_sim
start_cm

echo "Idling..."
sleep 10

create_shared_instance

run_tests

delete_shared_instance

echo "Done :)"
