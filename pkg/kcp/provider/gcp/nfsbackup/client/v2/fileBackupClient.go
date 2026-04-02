package v2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// FileBackupClient defines operations for GCP Filestore backups.
// This interface abstracts the GCP Filestore Backup API for easier testing and mocking.
type FileBackupClient interface {
	// GetBackup retrieves a Filestore backup by name.
	GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error)

	// ListBackups lists Filestore backups with optional filter.
	ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error)

	// CreateBackup creates a new Filestore backup.
	// Returns the operation name for tracking.
	CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error)

	// DeleteBackup deletes a Filestore backup.
	// Returns the operation name for tracking.
	DeleteBackup(ctx context.Context, projectId, location, name string) (string, error)

	// GetBackupLROperation retrieves the status of a long-running operation.
	// Returns the full operation object to allow callers to check Done status and Error.
	GetBackupLROperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error)

	// UpdateBackup updates an existing Filestore backup (e.g., labels).
	// Returns the operation name for tracking.
	UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error)
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
		filestoreClient: filestoreClient,
	}
}

// fileBackupClient implements FileBackupClient using the wrapped gcpclient.FilestoreClient.
type fileBackupClient struct {
	filestoreClient gcpclient.FilestoreClient
}

var _ FileBackupClient = &fileBackupClient{}

func (c *fileBackupClient) GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "name", name)

	req := &filestorepb.GetBackupRequest{
		Name: GetFileBackupPath(projectId, location, name),
	}

	backup, err := c.filestoreClient.GetFilestoreBackup(ctx, req)
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Filestore backup not found")
			return nil, err
		}
		logger.Error(err, "Failed to get Filestore backup")
		return nil, err
	}

	return backup, nil
}

func (c *fileBackupClient) ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error) {
	logger := composed.LoggerFromCtx(ctx)

	// Use "-" for all locations
	req := &filestorepb.ListBackupsRequest{
		Parent: GetFilestoreParentPath(projectId, "-"),
		Filter: filter,
	}

	var backups []*filestorepb.Backup
	it := c.filestoreClient.ListFilestoreBackups(ctx, req)
	for backup, err := range it.All() {
		if err != nil {
			logger.Error(err, "Failed to list backups", "projectId", projectId, "filter", filter)
			return nil, err
		}
		backups = append(backups, backup)
	}

	return backups, nil
}

func (c *fileBackupClient) CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.CreateBackupRequest{
		Parent:   GetFilestoreParentPath(projectId, location),
		BackupId: name,
		Backup:   backup,
	}

	op, err := c.filestoreClient.CreateFilestoreBackup(ctx, req)
	if err != nil {
		logger.Error(err, "Failed to create Filestore backup",
			"projectId", projectId,
			"location", location,
			"name", name)
		return "", err
	}

	return op.Name(), nil
}

func (c *fileBackupClient) DeleteBackup(ctx context.Context, projectId, location, name string) (string, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "name", name)

	req := &filestorepb.DeleteBackupRequest{
		Name: GetFileBackupPath(projectId, location, name),
	}

	op, err := c.filestoreClient.DeleteFilestoreBackup(ctx, req)
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Filestore backup not found")
			return "", err
		}
		logger.Error(err, "Failed to delete Filestore backup")
		return "", err
	}

	return op.Name(), nil
}

func (c *fileBackupClient) GetBackupLROperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &longrunningpb.GetOperationRequest{
		Name: operationName,
	}

	op, err := c.filestoreClient.GetFilestoreOperation(ctx, req)
	if err != nil {
		logger.Error(err, "Failed to get operation", "operationName", operationName)
		return nil, err
	}

	return op, nil
}

func (c *fileBackupClient) UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	backup.Name = GetFileBackupPath(projectId, location, name)

	req := &filestorepb.UpdateBackupRequest{
		Backup: backup,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
	}

	op, err := c.filestoreClient.UpdateFilestoreBackup(ctx, req)
	if err != nil {
		logger.Error(err, "Failed to update Filestore backup",
			"projectId", projectId,
			"location", location,
			"name", name,
			"updateMask", updateMask)
		return "", err
	}

	return op.Name(), nil
}
