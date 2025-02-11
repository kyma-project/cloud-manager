package client

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
)

type FilestoreClient interface {
	GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)
	CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)
	DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)
	GetFilestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
	PatchFilestoreInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error)
}

func NewFilestoreClientProvider() client.ClientProvider[FilestoreClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (FilestoreClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, saJsonKeyPath)
			if err != nil {
				return nil, err
			}

			fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP File Client: [%w]", err)
			}
			return NewFilestoreClient(fsClient), nil
		},
	)
}

func NewFilestoreClient(svcFilestore *file.Service) FilestoreClient {
	return &filestoreClient{svcFilestore: svcFilestore}
}

type filestoreClient struct {
	svcFilestore *file.Service
}

func (c *filestoreClient) GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcFilestore.Projects.Locations.Instances.Get(client.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	client.IncrementCallCounter("File", "Instances.Get", location, err)
	if err != nil {
		logger.Info("GetFilestoreInstance", "err", err)
		return nil, err
	}
	return out, err
}

func (c *filestoreClient) CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Create(client.GetFilestoreParentPath(projectId, location), instance).InstanceId(instanceId).Do()
	client.IncrementCallCounter("File", "Instances.Create", location, err)
	if err != nil {
		logger.Error(err, "CreateFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Delete(client.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	client.IncrementCallCounter("File", "Instances.Delete", location, err)
	if err != nil {
		logger.Error(err, "DeleteFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) GetFilestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Operations.Get(operationName).Do()
	client.IncrementCallCounter("File", "Operations.Get", "", err)
	if err != nil {
		logger.Error(err, "GetFilestoreOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return operation, nil
}

// PatchFilestoreInstance updates the Filestore instance.
// UpdateMask is a comma-separated list of fully qualified names of fields that should be updated in this request.
func (c *filestoreClient) PatchFilestoreInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Patch(client.GetFilestoreInstancePath(projectId, location, instanceId), instance).UpdateMask(updateMask).Do()
	client.IncrementCallCounter("File", "Instances.Patch", location, err)
	if err != nil {
		logger.Error(err, "PatchFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}
