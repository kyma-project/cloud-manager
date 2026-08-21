package mock

import (
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	alicloudredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	alicloudredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

// VpcConfig is the test-side seeding API for VPCs and vSwitches.
type VpcConfig interface {
	AddVpc(id, name, cidr string) *VpcEntry
	AddVSwitch(vpcId, vSwitchId, name, zoneId, cidr string) *VSwitchEntry
	AddZone(zoneId string)
	SetVpcError(vpcId string, err error)
	SetVSwitchError(vSwitchId string, err error)
}

// NasConfig is the test-side seeding API for NAS file systems.
type NasConfig interface {
	AddNasFileSystem(id, protocolType, storageType, zoneId string) *NasFileSystemEntry
	SetNasFileSystemError(fileSystemId string, err error)
}

// RedisInstanceConfig is the test-side seeding API for AliCloud r-kvstore
// standard (non-sharded) HA instances.
type RedisInstanceConfig interface {
	AddRedisInstance(instanceId, instanceClass, engineVersion, status string) *RedisInstanceEntry
	SetRedisInstanceError(instanceId string, err error)
	// GetRedisInstance returns the stored entry for instanceId, or nil if not found.
	GetRedisInstance(instanceId string) *RedisInstanceEntry
}

// RedisClusterConfig is the test-side seeding API for AliCloud r-kvstore
// sharded cloud-native cluster instances. Kept as a distinct interface because
// the cluster entry carries additional shape (ShardCount, ReplicasPerShard).
type RedisClusterConfig interface {
	AddRedisCluster(instanceId, instanceClass, engineVersion, status string, shardCount int32) *RedisClusterEntry
	SetRedisClusterError(instanceId string, err error)
	// GetRedisCluster returns the stored entry for instanceId, or nil if not found.
	GetRedisCluster(instanceId string) *RedisClusterEntry
}

// Configs aggregates all test-side seeding interfaces.
type Configs interface {
	VpcConfig
	NasConfig
	RedisInstanceConfig
	RedisClusterConfig
}

// AccountRegion is the per-(account, region) mock surface.
type AccountRegion interface {
	Configs

	IpRangeClient() alicloudiprangeclient.Client
	VpcNetworkClient() alicloudvpcnetworkclient.Client
	NfsInstanceClient() alicloudnfsinstanceclient.Client
	RedisInstanceClient() alicloudredisinstanceclient.Client
	RedisClusterClient() alicloudredisclusterclient.Client

	Region() string

	// TransitionAllToNormal advances every in-flight r-kvstore entry
	// (Creating/Changing) to Normal. Test helper.
	TransitionAllToNormal()
}

// AccountCredential is the access-key pair for an account.
type AccountCredential struct {
	AccessKeyId     string
	AccessKeySecret string
}

// Account represents a single Alicloud account.
type Account interface {
	AccountId() string
	Credentials() AccountCredential
	Region(region string) AccountRegion
	Delete()
}

// Providers exposes ClientProvider funcs for controller suite wiring.
type Providers interface {
	IpRangeClientProvider() alicloudiprangeclient.ClientProvider
	VpcNetworkClientProvider() alicloudvpcnetworkclient.ClientProvider
	NfsInstanceClientProvider() alicloudnfsinstanceclient.ClientProvider
	RedisInstanceClientProvider() alicloudredisinstanceclient.ClientProvider
	RedisClusterClientProvider() alicloudredisclusterclient.ClientProvider
}

// Server is the top-level mock - owns accounts and yields providers.
type Server interface {
	Providers

	NewAccount() Account
	NewAccountWithCredentials(accessKeyId, accessKeySecret string) Account
	GetAccount(accountId string) Account
	Login(accessKeyId, accessKeySecret string) (Account, error)
}
