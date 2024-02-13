package client

import (
	"context"
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

func NewFilestoreClient() client.ClientProvider[FilestoreClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (FilestoreClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, saJsonKeyPath)
			if err != nil {
				return nil, err
			}

			fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
			return newFilestoreClient(fsClient), nil
		},
	)
}

func newFilestoreClient(svcFilestore *file.Service) FilestoreClient {
	return &filestoreClient{svcFilestore: svcFilestore}
}

type filestoreClient struct {
	svcFilestore *file.Service
}

func (c *filestoreClient) GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcFilestore.Projects.Locations.Instances.Get(client.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	if err != nil {
		logger.V(4).Info("GetFilestoreInstance", "err", err)
	}
	return out, err
}

func (c *filestoreClient) CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Create(client.GetFilestoreParentPath(projectId, location), instance).InstanceId(instanceId).Do()
	if err != nil {
		logger.Error(err, "CreateFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Delete(client.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	if err != nil {
		logger.Error(err, "DeleteFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) GetFilestoreOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Operations.Get(operationName).Do()
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
	if err != nil {
		logger.Error(err, "PatchFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}
