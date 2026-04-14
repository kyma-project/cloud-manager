package client

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// RegionOperationsClient embeds the wrapped gcpclient.ComputeRegionalOperationsClient interface.
// Actions call the wrapped methods directly (e.g., GetComputeRegionalOperation)
// by constructing the protobuf request inline.
type RegionOperationsClient interface {
	gcpclient.ComputeRegionalOperationsClient
}

func NewRegionOperationsClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[RegionOperationsClient] {
	return func(_ string) RegionOperationsClient {
		return NewRegionOperationsClient(gcpClients)
	}
}

func NewRegionOperationsClient(gcpClients *gcpclient.GcpClients) RegionOperationsClient {
	return NewRegionOperationsClientFromWrapped(gcpClients.RegionOperationsWrapped())
}

func NewRegionOperationsClientFromWrapped(regionalOpsClient gcpclient.ComputeRegionalOperationsClient) RegionOperationsClient {
	return &regionalOperationsClient{ComputeRegionalOperationsClient: regionalOpsClient}
}

type regionalOperationsClient struct {
	gcpclient.ComputeRegionalOperationsClient
}

var _ RegionOperationsClient = &regionalOperationsClient{}
