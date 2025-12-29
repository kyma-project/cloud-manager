// Package client provides the GCP Filestore client abstraction for v2.
package client

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/option"
)

// FilestoreClient defines the interface for interacting with GCP Filestore API.
// This interface abstracts the GCP API calls to enable easier testing and mocking.
type FilestoreClient interface {
	// GetInstance retrieves a Filestore instance by ID.
	GetInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)

	// CreateInstance creates a new Filestore instance.
	CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)

	// UpdateInstance updates an existing Filestore instance.
	// The updateMask parameter is a comma-separated list of fully qualified field names.
	UpdateInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error)

	// DeleteInstance deletes a Filestore instance.
	DeleteInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)

	// GetOperation retrieves the status of a long-running operation.
	GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
}

// filestoreClient implements FilestoreClient.
type filestoreClient struct {
	svcFilestore *file.Service
}

// NewFilestoreClient creates a new FilestoreClient.
func NewFilestoreClient(svcFilestore *file.Service) FilestoreClient {
	return &filestoreClient{svcFilestore: svcFilestore}
}

// NewFilestoreClientProvider creates a client provider for FilestoreClient.
func NewFilestoreClientProvider() gcpclient.ClientProvider[FilestoreClient] {
	return gcpclient.NewCachedClientProvider(
		func(ctx context.Context, credentialsFile string) (FilestoreClient, error) {
			httpClient, err := gcpclient.GetCachedGcpClient(ctx, credentialsFile)
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

// GetInstance retrieves a Filestore instance.
func (c *filestoreClient) GetInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error) {
	logger := composed.LoggerFromCtx(ctx)
	path := gcpclient.GetFilestoreInstancePath(projectId, location, instanceId)
	out, err := c.svcFilestore.Projects.Locations.Instances.Get(path).Do()
	gcpclient.IncrementCallCounter("File", "Instances.Get", location, err)
	if err != nil {
		logger.Info("GetInstance error", "err", err, "path", path)
		return nil, err
	}
	return out, nil
}

// CreateInstance creates a new Filestore instance.
func (c *filestoreClient) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	parent := gcpclient.GetFilestoreParentPath(projectId, location)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Create(parent, instance).InstanceId(instanceId).Do()
	gcpclient.IncrementCallCounter("File", "Instances.Create", location, err)
	if err != nil {
		logger.Error(err, "CreateInstance error", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

// UpdateInstance updates an existing Filestore instance.
func (c *filestoreClient) UpdateInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	path := gcpclient.GetFilestoreInstancePath(projectId, location, instanceId)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Patch(path, instance).UpdateMask(updateMask).Do()
	gcpclient.IncrementCallCounter("File", "Instances.Patch", location, err)
	if err != nil {
		logger.Error(err, "UpdateInstance error", "projectId", projectId, "location", location, "instanceId", instanceId, "updateMask", updateMask)
		return nil, err
	}
	return operation, nil
}

// DeleteInstance deletes a Filestore instance.
func (c *filestoreClient) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	path := gcpclient.GetFilestoreInstancePath(projectId, location, instanceId)
	operation, err := c.svcFilestore.Projects.Locations.Instances.Delete(path).Do()
	gcpclient.IncrementCallCounter("File", "Instances.Delete", location, err)
	if err != nil {
		logger.Error(err, "DeleteInstance error", "projectId", projectId, "location", location, "instanceId", instanceId)
		return nil, err
	}
	return operation, nil
}

// GetOperation retrieves a long-running operation.
func (c *filestoreClient) GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcFilestore.Projects.Locations.Operations.Get(operationName).Do()
	gcpclient.IncrementCallCounter("File", "Operations.Get", "", err)
	if err != nil {
		logger.Error(err, "GetOperation error", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return operation, nil
}
