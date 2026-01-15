package mock

import (
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	awsvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcnetwork/client"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"
)

type IpRangeClient interface {
	awsiprangeclient.Client
}

type VpcPeeringClient interface {
	awsvpcpeeringclient.Client
}

type NfsClient interface {
	awsnfsinstanceclient.Client
}

type ElastiCacheClient interface {
	awsclient.ElastiCacheClient
}

type ScopeClient interface {
	scopeclient.AwsStsClient
}

type SubscriptionClient interface {
	subscriptionclient.AwsStsClient
}

type ExposedDataClient interface {
	awsexposeddataclient.Client
}

type VpcNetworkClient interface {
	awsvpcnetworkclient.Client
}

type Clients interface {
	IpRangeClient
	NfsClient
	VpcPeeringClient
	ElastiCacheClient
	ExposedDataClient
	VpcNetworkClient
}

type Providers interface {
	ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient]
	SubscriptionGardenProvider() awsclient.GardenClientProvider[subscriptionclient.AwsStsClient]
	IpRangeSkrProvider() awsclient.SkrClientProvider[awsiprangeclient.Client]
	NfsInstanceSkrProvider() awsclient.SkrClientProvider[awsnfsinstanceclient.Client]
	VpcPeeringSkrProvider() awsclient.SkrClientProvider[awsvpcpeeringclient.Client]
	ElastiCacheProviderFake() awsclient.SkrClientProvider[awsclient.ElastiCacheClient]
	ExposedDataProvider() awsclient.SkrClientProvider[awsexposeddataclient.Client]
	VpcNetworkProvider() awsclient.SkrClientProvider[awsvpcnetworkclient.Client]
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

	Region() string
}

type Server interface {
	Providers

	NewAccount() Account
	GetAccount(accountId string) Account
	Login(key, secret string) (Account, error)
}
