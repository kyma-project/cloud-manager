package client

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type Client interface {
	gcpclient.SecurityCenterManagementClient
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
	return func(_ string) Client {
		return &client{
			SecurityCenterManagementClient: gcpClients.SecurityCenterManagementWrapped(),
		}
	}
}

type client struct {
	gcpclient.SecurityCenterManagementClient
}
