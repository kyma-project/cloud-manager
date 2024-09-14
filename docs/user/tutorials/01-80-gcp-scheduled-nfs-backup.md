# Create scheduled automatic RWX Volume backups in GCP

This tutorial explains how to create scheduled automatic backups for ReadWriteMany (RWX) volumes in Google Cloud Platform (GCP).

## Preparation <!-- {docsify-ignore} -->

Create a GcpNfsVolume by referring to [Use RWX Volumes in GCP](./01-20-gcp-nfs-volume.md).

All the examples below assume that the GcpNfsVolume is named `my-vol` and is in the same namespace as the GcpNfsBackupSchedule resource.

## Steps <!-- {docsify-ignore} -->

1. Export the namespace as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   ```

2. Create a GcpNfsBackupSchedule resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsBackupSchedule
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

3. Wait for the GcpNfsVolumeBackup to be in the `Active` state.
   ```shell
   kubectl -n $NAMESPACE wait --for=jsonpath='{.status.state}'=Active gcpnfsbackupschedule/my-backup-schedule --timeout=300s
   ```
   Once the GcpNfsVolumeBackup is created, you should see the following message:
   ```
   gcpnfsbackupschedule.cloud-resources.kyma-project.io/my-backup-schedule condition met
   ```
4. Observe the nextRunTimes for the creating the backups.
   ```shell
   kubectl -n $NAMESPACE get gcpnfsbackupschedule my-backup-schedule -o jsonpath='{.status.nextRunTimes}{"\n"}' 
   ```
5. Wait till the time specified in the nextRunTimes (in previous step) and see that the GcpNfsVolumeBackup objects get created.
   ```shell
   kubectl -n $NAMESPACE get gcpnfsvolumebackup -l cloud-resources.kyma-project.io/scheduleName=my-backup-schedule 
   ```
6. Clean up:
    * Remove the created schedule and the backups:
      ```shell
      kubectl delete -n $NAMESPACE gcpnfsbackupschedule my-backup-schedule
      ```