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
		return NewComputeClient(gcpClients)
	}
}

func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
	return NewComputeClientFromSubnetClient(gcpClients.SubnetWrapped())
}

func NewComputeClientFromSubnetClient(subnetClient gcpclient.SubnetClient) ComputeClient {
	return &computeClient{SubnetClient: subnetClient}
}

type computeClient struct {
	gcpclient.SubnetClient
}

var _ ComputeClient = &computeClient{}
