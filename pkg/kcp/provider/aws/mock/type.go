package mock

import (
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
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

type ScopeClient interface {
	scopeclient.AwsStsClient
}

type Clients interface {
	IpRangeClient
	NfsClient
	VpcPeeringClient
	awsclient.ElastiCacheClient
}

type Providers interface {
	ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient]
	IpRangeSkrProvider() awsclient.SkrClientProvider[awsiprangeclient.Client]
	NfsInstanceSkrProvider() awsclient.SkrClientProvider[awsnfsinstanceclient.Client]
	VpcPeeringSkrProvider() awsclient.SkrClientProvider[awsvpcpeeringclient.Client]
	ElastiCacheProviderFake() awsclient.SkrClientProvider[awsclient.ElastiCacheClient]
	ExposedDataProvider() awsclient.SkrClientProvider[awsexposeddataclient.Client]
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
