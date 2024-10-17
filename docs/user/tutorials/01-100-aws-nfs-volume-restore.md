# Restore RWX Volume Backups in AWS

This tutorial explains how to initiate a restore operation for the ReadWriteMany (RWX) volumes in Amazon Web Services (AWS). 

## Prerequisites <!-- {docsify-ignore} -->

* You have created an AwsNfsVolume. See [Use RWX Volumes in AWS](./01-10-aws-nfs-volume.md) to learn more.

* You have created an AwsNfsVolumeBackup. See [Backup RWX Volumes in AWS](./01-110-aws-nfs-volume-backup.md) to learn more.

>[!NOTE]
>All the examples below assume that the AwsNfsVolumeBackup is named `my-backup` 
and is in the same namespace as the AwsNfsVolumeRestore resource.

## Steps <!-- {docsify-ignore} -->

### Restore on the same or existing Filestore <!-- {docsify-ignore} -->
1. Export the namespace as an environment variable.

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```
   
2. Create an AwsNfsVolumeRestore resource. 

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsVolumeRestore
   metadata:
     name: my-restore
   spec:
     source:
       backup:
         name: my-backup
   EOF
   ```
   
3. Wait for the AwsNfsVolumeRestore to be in the `Done` state and have the `Ready` condition.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready awsnfsvolumerestore/my-restore --timeout=600s
   ```

   Once the AwsNfsVolumeRestore is completed, you should see the following message:

   ```
   awsnfsvolumerestore.cloud-resources.kyma-project.io/my-restore condition met
   ```

4. Clean up:

   Remove the created awsnfsvolumeRestore:
   ```shell
   kubectl delete -n $NAMESPACE awsnfsvolumerestore my-restore
   ```