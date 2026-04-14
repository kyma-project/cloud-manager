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
		return gcpClients.FilestoreWrapped()
	}
}
