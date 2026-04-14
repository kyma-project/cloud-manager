package client

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// FilestoreClient embeds the wrapped gcpclient.FilestoreClient interface.
// Actions call the wrapped methods directly (e.g., GetFilestoreInstance, CreateFilestoreInstance)
// using name-building utilities from util.go.
type FilestoreClient interface {
	gcpclient.FilestoreClient
}

// NewFilestoreClientProvider creates a provider function for FilestoreClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFilestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FilestoreClient] {
	return func(_ string) FilestoreClient {
		return NewFilestoreClient(gcpClients)
	}
}

// NewFilestoreClient creates a new FilestoreClient wrapping GcpClients.
func NewFilestoreClient(gcpClients *gcpclient.GcpClients) FilestoreClient {
	return NewFilestoreClientFromFilestoreClient(gcpClients.FilestoreWrapped())
}

func NewFilestoreClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FilestoreClient {
	return &filestoreClientImpl{
		FilestoreClient: filestoreClient,
	}
}

// filestoreClientImpl implements FilestoreClient by embedding the wrapped GCP Filestore interface.
type filestoreClientImpl struct {
	gcpclient.FilestoreClient
}

var _ FilestoreClient = &filestoreClientImpl{}
