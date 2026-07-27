package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
)

type IpRangeClient interface {
	gcpiprangeclient.ComputeClient
	gcpiprangeclient.ServiceNetworkingClient
}

type Clients interface {
	IpRangeClient
	gcpexposeddataclient.Client
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	// GcpClientProvider versions for NEW pattern
	ServiceNetworkingClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	ComputeClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ComputeClient]
	ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient]
	OldComputeClientProvider() client.ClientProvider[gcpiprangeclient.OldComputeClient] // For v2 legacy
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client]
}

type Server interface {
	Clients

	Providers

	IpRangeClientUtils

	VpcPeeringMockClientUtils

	ExposedDataConfig
}
