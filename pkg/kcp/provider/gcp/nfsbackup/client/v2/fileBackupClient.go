package v2

import (
	"context"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/api/iterator"
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

	// GetOperation retrieves the status of a long-running operation.
	// Returns true if the operation is done, false otherwise.
	GetOperation(ctx context.Context, operationName string) (bool, error)

	// UpdateBackup updates an existing Filestore backup (e.g., labels).
	// Returns the operation name for tracking.
	UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error)
}

// NewFileBackupClientProvider creates a provider function for FileBackupClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFileBackupClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileBackupClient] {
	return func() FileBackupClient {
		return NewFileBackupClient(gcpClients)
	}
}

// NewFileBackupClient creates a new FileBackupClient wrapping GcpClients.
func NewFileBackupClient(gcpClients *gcpclient.GcpClients) FileBackupClient {
	return &fileBackupClient{
		cloudFilestoreManager: gcpClients.Filestore,
	}
}

// fileBackupClient implements FileBackupClient using the modern GCP Filestore API.
type fileBackupClient struct {
	cloudFilestoreManager *filestore.CloudFilestoreManagerClient
}

var _ FileBackupClient = &fileBackupClient{}

func (c *fileBackupClient) GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "name", name)

	req := &filestorepb.GetBackupRequest{
		Name: GetFileBackupPath(projectId, location, name),
	}

	backup, err := c.cloudFilestoreManager.GetBackup(ctx, req)
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
	it := c.cloudFilestoreManager.ListBackups(ctx, req)
	for {
		backup, err := it.Next()
		if err == iterator.Done {
			break
		}
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

	op, err := c.cloudFilestoreManager.CreateBackup(ctx, req)
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

	op, err := c.cloudFilestoreManager.DeleteBackup(ctx, req)
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

func (c *fileBackupClient) GetOperation(ctx context.Context, operationName string) (bool, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &longrunningpb.GetOperationRequest{
		Name: operationName,
	}

	op, err := c.cloudFilestoreManager.LROClient.GetOperation(ctx, req)
	if err != nil {
		logger.Error(err, "Failed to get operation", "operationName", operationName)
		return false, err
	}

	return op.Done, nil
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

	op, err := c.cloudFilestoreManager.UpdateBackup(ctx, req)
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
