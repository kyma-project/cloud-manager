#!/usr/bin/env bash

set -e
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

rm -f $SCRIPT_DIR/crd/bases/_.yaml

echo "Patching CRDs..."

yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.1.1"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.4"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumes.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.6"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumebackups.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.19"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisinstances.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.11"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.20"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisinstances.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.4"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisclusters.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.1"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpsubnets.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.4"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurevpcpeerings.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.56"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azureredisinstances.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.4"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azureredisclusters.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.7"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.4"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumerestores.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.7"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsbackupschedules.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.3"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpvpcpeerings.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.3"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsvpcpeerings.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.1"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_cloudresources.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.5"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsbackupschedules.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxvolumebackups.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxvolumerestores.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.3"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxbackupschedules.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.3"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisclusters.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurevpcdnslinks.yaml
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.1"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_sapnfsvolumes.yaml
