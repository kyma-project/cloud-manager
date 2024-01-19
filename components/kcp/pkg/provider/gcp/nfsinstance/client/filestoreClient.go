package client

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"net/http"

	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
)

type FilestoreClient interface {
	GetFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)
	CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)
	DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)
	GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
}

func NewFilestoreClient() gcpclient.ClientProvider[FilestoreClient] {
	return gcpclient.NewCachedClientProvider(
		func(ctx context.Context, httpClient *http.Client) (FilestoreClient, error) {
			client, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
			return newFilestoreClient(client), nil
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
	out, err := c.svcFilestore.Projects.Locations.Instances.Get(gcpclient.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	if err != nil {
		logger.V(4).Info("GetFilestoreInstance", "err", err)
	}
	return out, err
}

func (c *filestoreClient) CreateFilestoreInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Create(gcpclient.GetFilestoreInstancePath(projectId, location, instanceId), instance).Do()
	if err != nil {
		logger.Error(err, "CreateFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) DeleteFilestoreInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Delete(gcpclient.GetFilestoreInstancePath(projectId, location, instanceId)).Do()
	if err != nil {
		logger.Error(err, "DeleteFilestoreInstance", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

func (c *filestoreClient) GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Operations.Get(operationName).Do()
	if err != nil {
		logger.Error(err, "GetOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
