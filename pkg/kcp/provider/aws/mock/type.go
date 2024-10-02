package mock

import (
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scope "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

type IpRangeClient interface {
	iprangeclient.Client
}

type VpcPeeringClient interface {
	vpcpeeringclient.Client
}

type NfsClient interface {
	nfsinstanceclient.Client
}

type ScopeClient interface {
	scope.AwsStsClient
}

type Clients interface {
	IpRangeClient
	NfsClient
	VpcPeeringClient
	redisinstanceclient.ElastiCacheClient
}

type Providers interface {
	ScopeGardenProvider() awsclient.GardenClientProvider[scope.AwsStsClient]
	IpRangeSkrProvider() awsclient.SkrClientProvider[iprangeclient.Client]
	NfsInstanceSkrProvider() awsclient.SkrClientProvider[nfsinstanceclient.Client]
	VpcPeeringSkrProvider() awsclient.SkrClientProvider[vpcpeeringclient.Client]
	ElastiCacheProviderFake() awsclient.SkrClientProvider[redisinstanceclient.ElastiCacheClient]
}

type Configs interface {
	VpcConfig
	NfsConfig
	VpcPeeringConfig
	RouteTableConfig
	AwsElastiCacheMockUtils
}

type AccountRegion interface {
	Clients
	Configs
}
type Server interface {
	Providers

	ScopeClient
	ScopeConfig

	MockConfigs(account, region string) AccountRegion
}
