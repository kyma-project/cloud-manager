package mock

import (
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

type Providers interface {
	VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client]

	MemoryStoreProvider() provider.SkrClientProvider[azureredisinstanceclient.MemorystoreClient]
}

type Server interface {
	Providers

	VpcNetworkConfig
}
