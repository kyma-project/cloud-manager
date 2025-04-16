#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# KCP =============================================================

# CRD
find $SCRIPT_DIR/crd/bases -type f -iname "cloud-control*.yaml" -exec cp "{}" $SCRIPT_DIR/dist/kcp/crd/bases \;

# SKR =============================================================

# Common
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_cloudresources.yaml $SCRIPT_DIR/dist/skr/crd/bases
rm -f $SCRIPT_DIR/dist/skr/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml 2> /dev/null

# ============= AWS ================

# AWS
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumes.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsvpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsredisclusters.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumebackups.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsbackupschedules.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_awsnfsvolumerestores.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws

# AWS UI
cp $SCRIPT_DIR/ui-extensions/awsnfsvolumes/cloud-resources.kyma-project.io_awsnfsvolumes_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsredisinstances/cloud-resources.kyma-project.io_awsredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsvpcpeerings/cloud-resources.kyma-project.io_awsvpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsnfsvolumebackups/cloud-resources.kyma-project.io_awsnfsvolumebackups_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsnfsvolumerestores/cloud-resources.kyma-project.io_awsnfsvolumerestores_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws
cp $SCRIPT_DIR/ui-extensions/awsnfsbackupschedules/cloud-resources.kyma-project.io_awsnfsbackupschedules_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/aws

# ============= GCP ================

# GCP
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumerestores.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpvpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpsubnets.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsbackupschedules.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp

# GCP UI
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumes/cloud-resources.kyma-project.io_gcpnfsvolumes_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumebackups/cloud-resources.kyma-project.io_gcpnfsvolumebackups_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsvolumerestores/cloud-resources.kyma-project.io_gcpnfsvolumerestores_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpvpcpeerings/cloud-resources.kyma-project.io_gcpvpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpredisinstances/cloud-resources.kyma-project.io_gcpredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/ui-extensions/gcpnfsbackupschedules/cloud-resources.kyma-project.io_gcpnfsbackupschedules_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp

# ============= AZURE ================

# AZURE
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurevpcpeerings.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/cloud-resources.kyma-project.io_azurevpcpeerings.yaml
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azureredisinstances.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/cloud-resources.kyma-project.io_azureredisinstances.yaml
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azureredisclusters.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/cloud-resources.kyma-project.io_azureredisclusters.yaml
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxvolumebackups.yaml    $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxvolumerestores.yaml    $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_azurerwxbackupschedules.yaml    $SCRIPT_DIR/dist/skr/crd/bases/providers/azure/

# AZURE UI
cp $SCRIPT_DIR/ui-extensions/azurevpcpeerings/cloud-resources.kyma-project.io_azurevpcpeerings_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure
cp $SCRIPT_DIR/ui-extensions/azureredisinstances/cloud-resources.kyma-project.io_azureredisinstances_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure
cp $SCRIPT_DIR/ui-extensions/ipranges/cloud-resources.kyma-project.io_ipranges_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure
cp $SCRIPT_DIR/ui-extensions/azurerwxbackupschedules/cloud-resources.kyma-project.io_azurerwxbackupschedules_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure
cp $SCRIPT_DIR/ui-extensions/azurerwxvolumerestores/cloud-resources.kyma-project.io_azurerwxvolumerestores_ui.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure


# ============= CCEE ================

# CCEE
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_cceenfsvolumes.yaml  $SCRIPT_DIR/dist/skr/crd/bases/providers/openstack


echo "CRD resources are copied to ./dist kcp and skr dirs"
echo "Note that no files are removed - you must remove them manually"
echo "Don't forget to adjust the kustomization.yaml files in that case as well!!!"
