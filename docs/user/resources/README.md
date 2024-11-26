# Cloud Manager Resources

The API of the Cloud Manager module is based on Kubernetes Custom Resource Definitions (CRD), which extend the Kubernetes API with custom additions. To inspect the specification of the Cloud Manager module API, see the following custom resources (CRs):

<!-- The table of contents is not part of the Help Portal docu.-->
* IP Range
  * [IpRange CR](#iprange-cr)
* NFS
  * [AwsNfsVolume CR](#awsnfsvolume-cr)
  * [AwsNfsVolumeBackup CR](#awsnfsvolumebackup-cr)
  * [AwsNfsBackupSchedule CR](#awsnfsbackupschedule-cr)
  * [AwsNfsVolumeRestore CR](#awsnfsvolumerestore-cr)
  * [GcpNfsVolume CR](#gcpnfsvolume-cr)
  * [GcpNfsVolumeBackup CR](#gcpnfsvolumebackup-cr)
  * [GcpNfsBackupSchedule CR](#gcpnfsbackupschedule-cr)
  * [GcpNfsVolumeRestore CR](#gcpnfsvolumerestore-cr)
* VPC Peering
  * [AzureVpcPeering CR](#azurevpcpeering-cr)
  * [AwsVpcPeering CR](#awsvpcpeering-cr)
  * [GcpVpcPeering CR](#gcpvpcpeering-cr)
* Redis  
  * [GcpRedisInstance CR](#gcpredisinstance-cr)
  * [AwsRedisInstance CR](#awsredisinstance-cr)
  * [AzureRedisInstance CR](#azureredisinstance-cr)

## IpRange CR

The `iprange.cloud-resources.kyma-project.io` CRD describes the Virtual Private Cloud (VPC) network IP range used for IP address allocation for cloud resources that require an IP address. For more information, see [IpRange Custom Resource](./04-10-iprange.md).

## AwsNfsVolume CR

The `awsnfsvolume.cloud-resources.kyma-project.io` CRD describes the Amazon Web Services Elastic File System (Amazon EFS) instance that can be used as a ReadWriteMany (RWX) volume in the cluster. For more information, see [AwsNfsVolume Custom Resource](./04-20-10-aws-nfs-volume.md).

## AwsNfsVolumeBackup CR

The `awsnfsvolumebackup.cloud-resources.kyma-project.io` CRD describes the backup of Amazon EFS. For more information, see [AwsNfsVolumeBackup Custom Resource](./04-20-11-aws-nfs-volume-backup.md).

## AwsNfsBackupSchedule CR

The `awsnfsbackupschedule.cloud-resources.kyma-project.io` CRD describes the user-defined schedule for creating a backup of AwsNfsVolume instances at regular intervals. For more information, see [AwsNfsBackupSchedule Custom Resource](./04-20-12-aws-nfs-backup-schedule.md).

## AwsNfsVolumeRestore CR

The `awsnfsvolumerestore.cloud-resources.kyma-project.io` CRD describes the Amazon EFS full restore operation on the existing EFS. For more information, see [AwsNfsVolumeRestore Custom Resource](./04-20-13-aws-nfs-volume-restore.md).

## GcpNfsVolume CR

The `gcpnfsvolume.cloud-resources.kyma-project.io` CRD describes the Google Cloud Filestore instance that can be used as an RWX volume in the cluster. For more information, see [GcpNfsVolume Custom Resource](./04-20-20-gcp-nfs-volume.md).

## GcpNfsVolumeBackup CR

The `gcpnfsvolumebackup.cloud-resources.kyma-project.io` CRD describes the backup of a Google Cloud Filestore instance. For more information, see [GcpNfsVolumeBackup Custom Resource](./04-20-21-gcp-nfs-volume-backup.md).

## GcpNfsBackupSchedule CR

The `gcpnfsbackupschedule.cloud-resources.kyma-project.io` CRD describes the backup of Google Cloud Filestore instances. For more information, see [GcpNfsBackupSchedule Custom Resource](./04-20-22-gcp-nfs-backup-schedule.md).

## GcpNfsVolumeRestore CR

The `gcpnfsvolumerestore.cloud-resources.kyma-project.io` CRD describes the Google Cloud Filestore restore operation. For more information, see [GcpNfsVolumeRestore Custom Resource](./04-20-23-gcp-nfs-volume-restore.md).

## AzureVpcPeering CR

The `azurevpcpeering.cloud-resources.kyma-project.io` CRD describes the Azure peering connection between Kyma and the remote Azure Virtual Network. For more information, see [AzureVpcPeering Custom Resource](./04-30-10-azure-vpc-peering.md).

## AwsVpcPeering CR

The `awsvpcpeering.cloud-resources.kyma-project.io` CRD describes the AWS peering connection between Kyma and the remote AWS Virtual Network. For more information, see [AwsVpcPeering Custom Resource](./04-30-20-aws-vpc-peering.md).

## GcpVpcPeering CR

The `gcpvpcpeering.cloud-resources.kyma-project.io` CRD describes the VPC peering that you can use to peer your Kyma cluster with your Google Cloud project VPC. For more information, see [GcpVpcPeering Custom Resource](./04-30-30-gcp-vpc-peering.md).

## GcpRedisInstance CR

The `gcpredisinstance.cloud-resources.kyma-project.io` CRD describes the Google Cloud Memorystore for Redis instance. For more information, see [GcpRedisInstance Custom Resource](./04-40-10-gcp-redis-instance.md).

## AwsRedisInstance CR

The `awsredisinstance.cloud-resources.kyma-project.io` CRD describes the Amazon ElastiCache Redis instance. For more information, see [AwsRedisInstance Custom Resource](./04-40-20-aws-redis-instance.md).

## AzureRedisInstance CR

The `azureredisinstance.cloud-resources.kyma-project.io` CRD describes the Microsoft Azure Cache for Redis instance. For more information, see [AzureRedisInstance Custom Resource](./04-40-30-azure-redis-instance.md).
