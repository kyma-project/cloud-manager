# Backup RWX Volumes in AWS

This tutorial explains how to create backups for ReadWriteMany (RWX) volumes in Google Cloud Platform (AWS). 

## Preparation <!-- {docsify-ignore} -->

You have created a AwsNfsVolume. Follow [Use RWX Volumes in AWS](./01-10-aws-nfs-volume.md) to learn more.

[!NOTE]
All the examples below assume that the AwsNfsVolume is named `my-vol` and is in the same namespace as the AwsNfsVolumeBackup resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```
   
2. Create an AwsNfsVolumeBackup resource. 

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsVolumeBackup
   metadata:
     name: my-backup
   spec:
     source:
       volume:
         name: my-vol
   EOF
   ```
   
3. Wait for the AwsNfsVolumeBackup to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready awsnfsvolumebackup/my-backup --timeout=300s
   ```

   Once the AwsNfsVolumeBackup is created, you should see the following message:

   ```
   awsnfsvolumebackup.cloud-resources.kyma-project.io/my-backup condition met
   ```
4. Observe the location of the created backup.

   ```shell
   kubectl -n $NAMESPACE get awsnfsvolumebackup my-backup -o jsonpath='{.status.location}{"\n"}' 
   ```

5. Clean up:

   * Remove the created awsnfsvolumebackup:
     ```shell
     kubectl delete -n $NAMESPACE awsnfsvolumebackup my-backup
     ```
