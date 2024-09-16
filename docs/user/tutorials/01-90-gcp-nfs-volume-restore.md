# Restore RWX Volume Backups in GCP

This tutorial explains how to initiate a restore operation for ReadWriteMany (RWX) volumes in Google Cloud Platform (GCP). 

## Prerequisites <!-- {docsify-ignore} -->

You have created a GcpNfsVolume. Follow [Use RWX Volumes in GCP](./01-20-gcp-nfs-volume.md) to learn more.

You have created a GcpNfsVolumeBackup. Follow[Backup RWX Volumes in GCP](./01-70-gcp-nfs-volume-backup.md) to learn more.

>[!NOTE]
>All the examples below assume that the GcpNfsVolume is named `my-vol`, the GcpNfsVolumeBackup is named `my-backup` 
and both are in the same namespace as the GcpNfsVolumeRestore resource.

## Steps <!-- {docsify-ignore} -->

### Restore on the same or existing Filestore <!-- {docsify-ignore} -->
1. Export the namespace as an environment variable.

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```
   
2. Create an GcpNfsVolumeRestore resource. 

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolumeRestore
   metadata:
     name: my-restore
   spec:
     source:
       backup:
         name: my-backup
         namespace: $NAMESPACE
     destination:
       volume:
         name: my-vol
         namespace: $NAMESPACE
   EOF
   ```
   
3. Wait for the GcpNfsVolumeRestore to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolumerestore/my-restore --timeout=600s
   ```

   Once the GcpNfsVolumeRestore is created, you should see the following message:

   ```
   gcpnfsvolumerestore.cloud-resources.kyma-project.io/my-restore condition met
   ```

4. Clean up:

   Remove the created gcpnfsvolumeRestore:
   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolumerestore my-restore
   ```
### Restore with creating a new Filestore <!-- {docsify-ignore} -->
1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create a new GcpNfsVolume resource with sourceBackup referring the existing backup.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolume
   metadata:
     name: my-vol2
   spec:
     location: us-west1-a
     capacityGb: 1024
     sourceBackup:
       name: my-backup
       namespace: $NAMESPACE
   EOF
   ```

3. Wait for the GcpNfsVolume to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolume/my-vol2 --timeout=600s
   ```

   Once the GcpNfsVolume is created, you should see the following message:

   ```
   gcpnfsvolume.cloud-resources.kyma-project.io/my-vol2 condition met
   ```
4. Clean up:
   Remove the created gcpnfsvolume:
   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolume my-vol2
   ```
>[!TIP]
>Optionally follow steps 4-8 from [Use RWX Volumes in GCP](./01-20-gcp-nfs-volume.md) using my-vol2 as GcpNfsVolume name to verify the created volume before clean up.
