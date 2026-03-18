package client

import gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

type Client interface {
	gcpclient.VpcNetworkClient
	gcpclient.RoutersClient
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
	return func(_ string) Client {
		return &client{
			VpcNetworkClient: gcpClients.NetworkWrapped(),
			RoutersClient:    gcpClients.RoutersWrapped(),
		}
	}
}

type client struct {
	gcpclient.VpcNetworkClient
	gcpclient.RoutersClient
}
