#!/usr/bin/env bash

# Error handling.
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# This scrpit generates the sec-scanners-config by fetching all relevant images.

TAG=$1
OUTPUT_FILE="sec-scanners-config-release.yaml"

# Generating File.
echo -e "generating to ${OUTPUT_FILE} \n"
cat <<EOF | tee "${OUTPUT_FILE}"
module-name: cloud-manager
kind: kyma
rc-tag: ${TAG}
bdba:
  - europe-docker.pkg.dev/kyma-project/prod/cloud-manager:${TAG}
mend:
  language: golang-mod
  exclude:
    - "**/*_test.go"
checkmarx-one:
  preset: go-default
  exclude:
    - '**/*_test.go'
    - 'pkg/testinfra/**'
EOF
