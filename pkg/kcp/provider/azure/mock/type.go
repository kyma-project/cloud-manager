package mock

import (
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

type Providers interface {
	VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client]
}

type Server interface {
	Providers

	VpcNetworkConfig
}
