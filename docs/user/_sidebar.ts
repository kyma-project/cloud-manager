export default [
  { text: 'Cloud Manager Module', link: './README' },
  { text: 'NFS', link: './00-20-nfs' },
  { text: 'VPC Peering', link: './00-30-vpc-peering', collapsed: true, items: [
    { text: 'Authorizing Cloud Manager in the Remote Cloud Provider', link: './00-31-vpc-peering-authorization' }
    ] },
  { text: 'Redis', link: './00-40-redis' },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'IpRange Custom Resource', link: './resources/04-10-iprange' },
    { text: 'AwsNfsVolume Custom Resource', link: './resources/04-20-10-aws-nfs-volume' },
    { text: 'AwsNfsVolumeBackup Custom Resource', link: './resources/04-20-11-aws-nfs-volume-backup' },
    { text: 'AwsNfsBackupSchedule Custom Resource', link: './resources/04-20-12-aws-nfs-backup-schedule' },
    { text: 'AwsNfsVolumeRestore Custom Resource', link: './resources/04-20-13-aws-nfs-volume-restore' },
    { text: 'GcpNfsVolume Custom Resource', link: './resources/04-20-20-gcp-nfs-volume' },
    { text: 'GcpNfsVolumeBackup Custom Resource', link: './resources/04-20-21-gcp-nfs-volume-backup' },
    { text: 'GcpNfsBackupSchedule Custom Resource', link: './resources/04-20-22-gcp-nfs-backup-schedule' },
    { text: 'GcpNfsVolumeRestore Custom Resource', link: './resources/04-20-23-gcp-nfs-volume-restore' },
    { text: 'AwsVpcPeering Custom Resource', link: './resources/04-30-10-aws-vpc-peering' },
    { text: 'GcpVpcPeering Custom Resource', link: './resources/04-30-20-gcp-vpc-peering' },
    { text: 'AzureVpcPeering Custom Resource', link: './resources/04-30-30-azure-vpc-peering.md' },
    { text: 'AwsRedisInstance Custom Resource', link: './resources/04-40-10-aws-redis-instance' },   
    { text: 'GcpRedisInstance Custom Resource', link: './resources/04-40-20-gcp-redis-instance.md' },
    { text: 'AzureRedisInstance Custom Resource', link: './resources/04-40-30-azure-redis-instance.md' },
    { text: 'AwsRedisCluster Custom Resource', link: './resources/04-50-10-azure-redis-cluster.md' },   
    { text: 'GcpRedisCluster Custom Resource', link: './resources/04-50-20-gcp-redis-cluster' },
    { text: 'GcpSubnet Custom Resource', link: './resources/04-50-21-gcp-subnet' },
    { text: 'AzureRedisCluster Custom Resource', link: './resources/04-50-30-azure-redis-cluster' }
    ] },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Using NFS in Amazon Web Services', link: './tutorials/01-20-10-aws-nfs-volume' },
    { text: 'Using NFS in Google Cloud', link: './tutorials/01-20-20-gcp-nfs-volume' },
    { text: 'Creating VPC Peering in Amazon Web Services', link: './tutorials/01-30-10-aws-vpc-peering' },
    { text: 'Creating VPC Peering in Google Cloud', link: './tutorials/01-30-20-gcp-vpc-peering' },
    { text: 'Creating VPC Peering in Microsoft Azure', link: './tutorials/01-30-30-azure-vpc-peering' },
    { text: 'Using AwsRedisInstance Custom Resources', link: './tutorials/01-40-10-aws-redis-instance' },
    { text: 'Using GcpRedisInstance Custom Resources', link: './tutorials/01-40-20-gcp-redis-instance' },
    { text: 'Using AzureRedisInstance Custom Resources', link: './tutorials/01-40-30-azure-redis-instance' },
    { text: 'Using AwsRedisCluster Custom Resources', link: './tutorials/01-50-10-aws-redis-cluster' },
    { text: 'Using GcpRedisCluster Custom Resources', link: './tutorials/01-50-20-gcp-redis-cluster' },
    { text: 'Using AzureRedisCluster Custom Resources', link: './tutorials/01-50-30-azure-redis-cluster' }
    ] },
  { text: 'Glossary', link: './00-10-glossary' }
];
