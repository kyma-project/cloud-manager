package v2

import (
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// FileBackupClient embeds the wrapped gcpclient.FilestoreClient interface.
// Actions call the wrapped methods directly (e.g., GetFilestoreBackup, CreateFilestoreBackup)
// using name-building utilities from util.go.
type FileBackupClient interface {
	gcpclient.FilestoreClient
}

// NewFileBackupClientProvider creates a provider function for FileBackupClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFileBackupClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileBackupClient] {
	return func(_ string) FileBackupClient {
		return NewFileBackupClient(gcpClients)
	}
}

// NewFileBackupClient creates a new FileBackupClient wrapping GcpClients.
func NewFileBackupClient(gcpClients *gcpclient.GcpClients) FileBackupClient {
	return NewFileBackupClientFromFilestoreClient(gcpClients.FilestoreWrapped())
}

// NewFileBackupClientFromFilestoreClient creates a FileBackupClient from a wrapped FilestoreClient.
// Used for mock2 wiring where the subscription's Store implements gcpclient.FilestoreClient.
func NewFileBackupClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FileBackupClient {
	return &fileBackupClient{
		FilestoreClient: filestoreClient,
	}
}

// fileBackupClient implements FileBackupClient by embedding the wrapped gcpclient.FilestoreClient.
type fileBackupClient struct {
	gcpclient.FilestoreClient
}

var _ FileBackupClient = &fileBackupClient{}
