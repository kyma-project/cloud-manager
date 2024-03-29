package mock

import (
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	scope "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

type IpRangeClient interface {
	iprangeclient.Client
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
	ScopeClient
}

type Providers interface {
	ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient]
	IpRangeSkrProvider() awsclient.SkrClientProvider[iprangeclient.Client]
	NfsInstanceSkrProvider() awsclient.SkrClientProvider[nfsinstanceclient.Client]
	//SkrProvider() awsclient.SkrClientProvider[Clients]
}

type Server interface {
	Clients

	Providers

	VpcConfig
	NfsConfig
	ScopeConfig
}
