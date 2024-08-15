#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# KCP =============================================================

# CRD
find $SCRIPT_DIR/crd/bases -type f -iname "cloud-control*.yaml" -exec cp "{}" $SCRIPT_DIR/dist/kcp/crd/bases \;

# SKR =============================================================

# Common
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_cloudresources.yaml $SCRIPT_DIR/dist/skr/crd/bases
rm -f $SCRIPT_DIR/dist/skr/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml 2> /dev/null

# AWS
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumes.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsvpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws

# GCP
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumerestores.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpvpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsbackupschedules.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp

# GCP UI
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumes/cloud-resources.kyma-project.io_gcpnfsvolumes_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumebackups/cloud-resources.kyma-project.io_gcpnfsvolumebackups_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumerestores/cloud-resources.kyma-project.io_gcpnfsvolumerestores_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpvpcpeerings/cloud-resources.kyma-project.io_gcpvpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpredisinstances/cloud-resources.kyma-project.io_gcpredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsbackupschedules/cloud-resources.kyma-project.io_gcpnfsbackupschedules_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp


# AWS UI
cp $SCRIPT_DIR/ui-extensions/awsnfsvolumes/cloud-resources.kyma-project.io_awsnfsvolumes_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsredisinstances/cloud-resources.kyma-project.io_awsredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsvpcpeerings/cloud-resources.kyma-project.io_awsvpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws

# AZURE
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurevpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/cloud-resources.kyma-project.io_azurevpcpeerings.yaml
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azureredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/cloud-resources.kyma-project.io_azureredisinstances.yaml

# AZURE UI
cp $SCRIPT_DIR/ui-extensions/azurevpcpeerings/cloud-resources.kyma-project.io_azurevpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure
cp $SCRIPT_DIR/ui-extensions/azureredisinstances/cloud-resources.kyma-project.io_azureredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure




echo "CRD resources are copied to ./dist kcp and skr dirs"
echo "Note that no files are removed - you must remove them manually"
echo "Don't forget to adjust the kustomization.yaml files in that case as well!!!"
