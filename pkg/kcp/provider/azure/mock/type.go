package mock

import (
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

type VpcPeeringClient interface {
	client.Client
}

type Clients interface {
	VpcPeeringClient
}

type Providers interface {
	VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client]
}

type Server interface {
	Clients

	Providers

	VpcPeeringConfig
}
