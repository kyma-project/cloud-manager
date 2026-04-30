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
		return gcpClients.RegionOperationsWrapped()
	}
}
