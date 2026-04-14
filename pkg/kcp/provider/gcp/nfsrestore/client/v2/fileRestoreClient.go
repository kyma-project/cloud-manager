package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// FileRestoreClient embeds the wrapped gcpclient.FilestoreClient interface and adds
// the FindRestoreOperation method which contains real business logic (filter construction,
// iteration to find running operations).
// Actions call the wrapped methods directly (e.g., RestoreFilestoreInstance, GetFilestoreOperation)
// using name-building utilities from util.go.
type FileRestoreClient interface {
	gcpclient.FilestoreClient

	// FindRestoreOperation lists operations for an instance to find a running restore.
	FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error)
}

// NewFileRestoreClientProvider creates a provider function for FileRestoreClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFileRestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileRestoreClient] {
	return func(_ string) FileRestoreClient {
		return NewFileRestoreClientFromFilestoreClient(gcpClients.FilestoreWrapped())
	}
}

// NewFileRestoreClientFromFilestoreClient wraps a gcpclient.FilestoreClient into a FileRestoreClient.
// Cannot be eliminated because FileRestoreClient has a value-add method (FindRestoreOperation)
// beyond the embedded gcpclient.FilestoreClient, so a plain FilestoreClient does not satisfy
// the interface.
func NewFileRestoreClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FileRestoreClient {
	return &fileRestoreClient{
		FilestoreClient: filestoreClient,
	}
}

// fileRestoreClient implements FileRestoreClient by embedding the wrapped gcpclient.FilestoreClient.
// Only FindRestoreOperation is kept as an additional method (value-add: filter construction + iteration).
type fileRestoreClient struct {
	gcpclient.FilestoreClient
}

var _ FileRestoreClient = &fileRestoreClient{}

func (c *fileRestoreClient) FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)

	filestoreParentPath := GetFilestoreParentPath(projectId, location)
	destFileFullPath := GetFilestoreInstancePath(projectId, location, instanceId)
	targetFilter := fmt.Sprintf("metadata.target=\"%s\"", destFileFullPath)
	verbFilter := "metadata.verb=\"restore\""
	filters := fmt.Sprintf("%s AND %s", targetFilter, verbFilter)

	it := c.ListFilestoreOperations(ctx, &longrunningpb.ListOperationsRequest{
		Name:   filestoreParentPath,
		Filter: filters,
	})
	var runningOperation *longrunningpb.Operation
	for op, err := range it.All() {
		if err != nil {
			logger.Error(err, "FindRestoreOperation",
				"projectId", projectId,
				"location", location,
				"instanceId", instanceId)
			return nil, err
		}
		if !op.Done {
			runningOperation = op
		}
	}

	if runningOperation == nil || runningOperation.Name == "" {
		return nil, nil
	}

	return runningOperation, nil
}
