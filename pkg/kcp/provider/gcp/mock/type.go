package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsinstancev1client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

type IpRangeClient interface {
	gcpiprangeclient.ComputeClient
	gcpiprangeclient.ServiceNetworkingClient
}

type NfsClient interface {
	gcpnfsinstancev1client.FilestoreClient
}

type Clients interface {
	IpRangeClient
	NfsClient
	gcpexposeddataclient.Client
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	// GcpClientProvider versions for NEW pattern
	ServiceNetworkingClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	ComputeClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ComputeClient]
	ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient]
	OldComputeClientProvider() client.ClientProvider[gcpiprangeclient.OldComputeClient] // For v2 legacy
	FilestoreClientProvider() client.ClientProvider[gcpnfsinstancev1client.FilestoreClient]
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client]
}

// ClientErrors is an interface for setting errors on the mock client to simulate Hyperscaler API errors
type ClientErrors interface {
	SetCreateError(err *googleapi.Error)
	SetPatchError(err *googleapi.Error)
	SetDeleteError(err *googleapi.Error)
	SetGetError(err *googleapi.Error)
	SetOperationError(err *googleapi.Error)
}

type Server interface {
	Clients

	Providers

	ClientErrors

	IpRangeClientUtils

	VpcPeeringMockClientUtils

	ExposedDataConfig
}
