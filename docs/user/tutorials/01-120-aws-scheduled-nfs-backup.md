# Create Scheduled Automatic RWX Volume Backups in AWS

This tutorial explains how to create scheduled automatic backups for ReadWriteMany (RWX) volumes in AWS.

## Prerequisites <!-- {docsify-ignore} -->

* You have created an AwsNfsVolume. See [Use RWX Volumes in AWS](./01-10-aws-nfs-volume.md) to learn more.

> [!NOTE]
> All the examples below assume that the AwsNfsVolume is named `my-vol` and is in the same namespace as the AwsNfsBackupSchedule resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create an AwsNfsBackupSchedule resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsBackupSchedule
   metadata:
      name: my-backup-schedule
   spec:
      nfsVolumeRef:
        name: my-vol
      schedule: "0 * * * *"
      prefix: my-hourly-backup
      deleteCascade: true
   EOF
   ```

3. Wait for the AwsNfsVolumeBackup to be in the `Active` state.
   ```shell
   kubectl -n $NAMESPACE wait --for=jsonpath='{.status.state}'=Active awsnfsbackupschedule/my-backup-schedule --timeout=300s
   ```
   Once the AwsNfsVolumeBackup is created, you should see the following message:
   ```
   awsnfsbackupschedule.cloud-resources.kyma-project.io/my-backup-schedule condition met
   ```
4. Observe the nextRunTimes for creating the backups.
   ```shell
   kubectl -n $NAMESPACE get awsnfsbackupschedule my-backup-schedule -o jsonpath='{.status.nextRunTimes}{"\n"}' 
   ```
5. Wait till the time specified in the nextRunTimes (in the previous step) passes and see that the AwsNfsVolumeBackup objects get created.
   ```shell
   kubectl -n $NAMESPACE get awsnfsvolumebackup -l cloud-resources.kyma-project.io/scheduleName=my-backup-schedule 
   ```
## Clean Up <!-- {docsify-ignore} -->
1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```
2. Clean up:
    * Remove the created schedule and the backups:
      ```shell
      kubectl delete -n $NAMESPACE awsnfsbackupschedule my-backup-schedule
      ```