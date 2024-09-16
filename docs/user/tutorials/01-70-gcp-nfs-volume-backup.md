# Backup RWX Volumes in GCP

This tutorial explains how to create backups for ReadWriteMany (RWX) volumes in Google Cloud Platform (GCP). 

## Preparation <!-- {docsify-ignore} -->

Create a GcpNfsVolume by referring to [Use RWX Volumes in GCP](./01-20-gcp-nfs-volume.md).

All the examples below assume that the GcpNfsVolume is named `my-vol` and is in the same namespace as the GcpNfsVolumeBackup resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```
   
2. Create an GcpNfsVolumeBackup resource. 

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolumeBackup
   metadata:
     name: my-backup
   spec:
     source:
       volume:
         name: my-vol
   EOF
   ```
   
3. Wait for the GcpNfsVolumeBackup to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolumebackup/my-backup --timeout=300s
   ```

   Once the GcpNfsVolumeBackup is created, you should see the following message:

   ```
   gcpnfsvolumebackup.cloud-resources.kyma-project.io/my-backup condition met
   ```
4. Observe the location of the created backup.

   ```shell
   kubectl -n $NAMESPACE get gcpnfsvolumebackup my-backup -o jsonpath='{.status.location}{"\n"}' 
   ```

5. Clean up:

   * Remove the created gcpnfsvolume:
     ```shell
     kubectl delete -n $NAMESPACE gcpnfsvolumebackup my-backup
     ```
