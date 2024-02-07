package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	nfsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
)

type IpRangeClient interface {
	iprangeclient.ComputeClient
	iprangeclient.ServiceNetworkingClient
}

type NfsClient interface {
	nfsclient.FilestoreClient
}

type Clients interface {
	IpRangeClient
	NfsClient
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[iprangeclient.ServiceNetworkingClient]
	ComputeClientProvider() client.ClientProvider[iprangeclient.ComputeClient]
	FilestoreClientProvider() client.ClientProvider[nfsclient.FilestoreClient]
}

type Server interface {
	Clients

	Providers
}
