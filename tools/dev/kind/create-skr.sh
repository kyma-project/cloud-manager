#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR=`realpath "$SCRIPT_DIR/../../.."`
export NO="$1"

kind create cluster --config - <<EOF
`envsubst < $PROJECT_ROOT_DIR/tools/dev/kind/kind-skr-config.yaml`
EOF

kind export kubeconfig -n skr$NO --kubeconfig "$PROJECT_ROOT_DIR/tools/dev/kind/kubeconfig-skr$NO.yaml"
