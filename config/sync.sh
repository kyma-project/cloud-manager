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

# GCP
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_ipranges.yaml      $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp
cp $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/gcp

# AZURE
#cp $SCRIPT_DIR/crd/bases/some-file.yaml $SCRIPT_DIR/dist/skr/crd/bases/providers/azure




echo "CRD resources are copied to ./dist kcp and skr dirs"
echo "Note that no files are removed - you must remove them manually"
echo "Don't forget to adjust the kustomization.yaml files in that case as well!!!"
