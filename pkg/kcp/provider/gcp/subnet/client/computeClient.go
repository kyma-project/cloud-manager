package client

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// ComputeClient embeds the wrapped gcpclient.SubnetClient interface.
// Actions call the wrapped methods directly (e.g., InsertSubnet, GetSubnet, DeleteSubnet)
// by constructing the protobuf request inline.
type ComputeClient interface {
	gcpclient.SubnetClient
}

func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
	return func(_ string) ComputeClient {
		return gcpClients.SubnetWrapped()
	}
}
