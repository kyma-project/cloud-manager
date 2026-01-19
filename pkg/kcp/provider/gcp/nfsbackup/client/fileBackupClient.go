package client

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
)

type FileBackupClient interface {
	GetFileBackup(ctx context.Context, projectId, location, name string) (*file.Backup, error)
	ListFilesBackups(ctx context.Context, projectId, filter string) ([]*file.Backup, error)
	CreateFileBackup(ctx context.Context, projectId, location, name string, backup *file.Backup) (*file.Operation, error)
	DeleteFileBackup(ctx context.Context, projectId, location, name string) (*file.Operation, error)
	GetBackupOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
	PatchFileBackup(ctx context.Context, projectId, location, name, updateMask string, backup *file.Backup) (*file.Operation, error)
}

func NewFileBackupClientProvider() client.ClientProvider[FileBackupClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (FileBackupClient, error) {
			baseClient, err := client.GetCachedGcpClient(ctx, credentialsFile)
			if err != nil {
				return nil, err
			}

			httpClient := client.NewMetricsHTTPClient(baseClient.Transport)

			fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP File Client: [%w]", err)
			}
			return NewFileBackupClient(fsClient), nil
		},
	)
}

func NewFileBackupClient(svcFile *file.Service) FileBackupClient {
	return &fileBackupClient{svcFile: svcFile}
}

type fileBackupClient struct {
	svcFile *file.Service
}

func (c *fileBackupClient) PatchFileBackup(ctx context.Context, projectId, location, name, updateMask string, backup *file.Backup) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFile.Projects.Locations.Backups.Patch(client.GetFileBackupPath(projectId, location, name), backup).UpdateMask(updateMask).Do()
	if err != nil {
		logger.Error(err, "PatchFileBackup", "projectId", projectId, "location", location, "name", name)
		return nil, err
	}
	return operation, nil
}

func (c *fileBackupClient) GetFileBackup(ctx context.Context, projectId, location, name string) (*file.Backup, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcFile.Projects.Locations.Backups.Get(client.GetFileBackupPath(projectId, location, name)).Do()
	if err != nil {
		logger.Info("GetFileBackup", "err", err)
		return nil, err
	}
	return out, err
}

func (c *fileBackupClient) ListFilesBackups(ctx context.Context, projectId, filter string) ([]*file.Backup, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcFile.Projects.Locations.Backups.List(client.GetFilestoreParentPath(projectId, "-")).Filter(filter).Do()
	if err != nil {
		logger.Info("ListFilesBackups", "err", err)
		return nil, err
	}
	return out.Backups, nil
}

func (c *fileBackupClient) CreateFileBackup(ctx context.Context, projectId, location, name string, backup *file.Backup) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFile.Projects.Locations.Backups.Create(client.GetFilestoreParentPath(projectId, location), backup).BackupId(name).Do()
	if err != nil {
		logger.Error(err, "CreateFileBackup", "projectId", projectId, "location", location, "name", name)
		return nil, err
	}
	return operation, nil
}
func (c *fileBackupClient) DeleteFileBackup(ctx context.Context, projectId, location, name string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFile.Projects.Locations.Backups.Delete(client.GetFileBackupPath(projectId, location, name)).Do()
	if err != nil {
		logger.Error(err, "DeleteFileBackup", "projectId", projectId, "location", location, "name", name)
		return nil, err
	}
	return operation, nil
}

func (c *fileBackupClient) GetBackupOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFile.Projects.Locations.Operations.Get(operationName).Do()
	if err != nil {
		logger.Error(err, "GetBackupOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
