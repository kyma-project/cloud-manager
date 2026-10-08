package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
)

type Clients interface {
	gcpexposeddataclient.Client
}

type Providers interface {
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client]
}

type Server interface {
	Clients

	Providers

	VpcPeeringMockClientUtils

	ExposedDataConfig
}
