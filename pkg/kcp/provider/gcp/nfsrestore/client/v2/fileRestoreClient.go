package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// FileRestoreClient defines operations for GCP Filestore restore.
// This interface abstracts the GCP Filestore Restore API for easier testing and mocking.
type FileRestoreClient interface {
	// RestoreFile triggers a restore from a backup to a Filestore instance.
	// Returns the operation name for tracking.
	RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (string, error)

	// GetRestoreOperation retrieves status of a long-running restore operation.
	// Returns the full operation object to allow callers to check Done status and Error.
	GetRestoreOperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error)

	// FindRestoreOperation lists operations for an instance to find a running restore.
	FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error)
}

// NewFileRestoreClientProvider creates a provider function for FileRestoreClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFileRestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileRestoreClient] {
	return func(_ string) FileRestoreClient {
		return NewFileRestoreClient(gcpClients)
	}
}

// NewFileRestoreClient creates a new FileRestoreClient wrapping GcpClients.
func NewFileRestoreClient(gcpClients *gcpclient.GcpClients) FileRestoreClient {
	return NewFileRestoreClientFromFilestoreClient(gcpClients.FilestoreWrapped())
}

// NewFileRestoreClientFromFilestoreClient creates a FileRestoreClient from an existing FilestoreClient.
// Used by mock2 to wire the subscription store into the feature client.
func NewFileRestoreClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FileRestoreClient {
	return &fileRestoreClient{
		filestoreClient: filestoreClient,
	}
}

// fileRestoreClient implements FileRestoreClient using the wrapped FilestoreClient interface.
type fileRestoreClient struct {
	filestoreClient gcpclient.FilestoreClient
}

var _ FileRestoreClient = &fileRestoreClient{}

func (c *fileRestoreClient) RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.RestoreInstanceRequest{
		Name:      destFileFullPath,
		FileShare: destFileShareName,
		Source: &filestorepb.RestoreInstanceRequest_SourceBackup{
			SourceBackup: srcBackupFullPath,
		},
	}

	op, err := c.filestoreClient.RestoreFilestoreInstance(ctx, req)
	if err != nil {
		logger.Error(err, "RestoreFile",
			"projectId", projectId,
			"destFileFullPath", destFileFullPath,
			"destFileShareName", destFileShareName,
			"srcBackupFullPath", srcBackupFullPath)
		return "", err
	}

	return op.Name(), nil
}

func (c *fileRestoreClient) GetRestoreOperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)

	op, err := c.filestoreClient.GetFilestoreOperation(ctx, &longrunningpb.GetOperationRequest{
		Name: operationName,
	})
	if err != nil {
		logger.Error(err, "GetRestoreOperation", "operationName", operationName)
		return nil, err
	}

	return op, nil
}

func (c *fileRestoreClient) FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)

	filestoreParentPath := GetFilestoreParentPath(projectId, location)
	destFileFullPath := GetFilestoreInstancePath(projectId, location, instanceId)
	targetFilter := fmt.Sprintf("metadata.target=\"%s\"", destFileFullPath)
	verbFilter := "metadata.verb=\"restore\""
	filters := fmt.Sprintf("%s AND %s", targetFilter, verbFilter)

	it := c.filestoreClient.ListFilestoreOperations(ctx, &longrunningpb.ListOperationsRequest{
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
