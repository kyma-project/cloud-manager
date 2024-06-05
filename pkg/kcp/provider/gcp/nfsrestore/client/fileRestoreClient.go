package client

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
)

type FileRestoreClient interface {
	RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (*file.Operation, error)
	GetRestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
}

func NewFileRestoreClientProvider() client.ClientProvider[FileRestoreClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (FileRestoreClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, saJsonKeyPath)
			if err != nil {
				return nil, err
			}

			fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP File Client: [%w]", err)
			}
			return NewFileRestoreClient(fsClient), nil
		},
	)
}

func NewFileRestoreClient(svcFile *file.Service) FileRestoreClient {
	return &fileRestoreClient{svcFile: svcFile}
}

type fileRestoreClient struct {
	svcFile *file.Service
}

func (c *fileRestoreClient) RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	request := &file.RestoreInstanceRequest{
		FileShare:    destFileShareName,
		SourceBackup: srcBackupFullPath,
	}
	operation, err := c.svcFile.Projects.Locations.Instances.Restore(destFileFullPath, request).Do()
	client.IncrementCallCounter("File", "Instances.Restore", "", err)
	if err != nil {
		logger.Error(err, "RestoreFile", "projectId", projectId, "destFileFullPath", destFileFullPath, "destFileShareName", destFileShareName, "srcBackupFullPath", srcBackupFullPath)
		return nil, err
	}
	return operation, nil
}

func (c *fileRestoreClient) GetRestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFile.Projects.Locations.Operations.Get(operationName).Do()
	client.IncrementCallCounter("File", "Operations.Get", "", err)
	if err != nil {
		logger.Error(err, "GetRestoreOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
