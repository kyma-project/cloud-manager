package mock

import (
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
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

// Configs aggregates all test-side seeding interfaces.
type Configs interface {
	VpcConfig
}

// AccountRegion is the per-(account, region) mock surface.
type AccountRegion interface {
	Configs

	IpRangeClient() alicloudiprangeclient.Client
	VpcNetworkClient() alicloudvpcnetworkclient.Client

	Region() string
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
}

// Server is the top-level mock — owns accounts and yields providers.
type Server interface {
	Providers

	NewAccount() Account
	NewAccountWithCredentials(accessKeyId, accessKeySecret string) Account
	GetAccount(accountId string) Account
	Login(accessKeyId, accessKeySecret string) (Account, error)
}
