#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR=`realpath "$SCRIPT_DIR/../../.."`

kind create cluster --config - <<EOF
`envsubst < $PROJECT_ROOT_DIR/tools/dev/kind/kind-kcp-config.yaml`
EOF

kind export kubeconfig -n kcp --kubeconfig $PROJECT_ROOT_DIR/tools/dev/kind/kubeconfig-kcp.yaml

