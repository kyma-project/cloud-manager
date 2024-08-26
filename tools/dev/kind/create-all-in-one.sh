#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR=`realpath "$SCRIPT_DIR/../../.."`

kind create cluster --config - <<EOF
`envsubst < $PROJECT_ROOT_DIR/tools/dev/kind/kind-all-in-one-config.yaml`
EOF

kind export kubeconfig -n kind --kubeconfig $PROJECT_ROOT_DIR/tools/dev/kind/kubeconfig-all-in-one.yaml

