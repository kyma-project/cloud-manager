package client

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// FilestoreClient defines business operations for GCP Filestore instances.
// This interface abstracts the GCP Filestore API for easier testing and mocking.
type FilestoreClient interface {
	// GetInstance retrieves a Filestore instance by name.
	GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error)

	// CreateInstance creates a new Filestore instance.
	// Returns the operation name for tracking.
	CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error)

	// UpdateInstance updates an existing Filestore instance.
	// The updateMask specifies which fields should be updated.
	// Returns the operation name for tracking.
	UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error)

	// DeleteInstance deletes a Filestore instance.
	// Returns the operation name for tracking.
	DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error)

	// GetOperation retrieves the status of a long-running operation.
	// Returns true if the operation is done, false otherwise.
	GetOperation(ctx context.Context, operationName string) (bool, error)
}

// NewFilestoreClientProvider creates a provider function for FilestoreClient instances.
// Follows the NEW pattern - accesses clients from GcpClients singleton.
func NewFilestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FilestoreClient] {
	return func(_ string) FilestoreClient {
		return NewFilestoreClient(gcpClients)
	}
}

// NewFilestoreClient creates a new FilestoreClient wrapping GcpClients.
func NewFilestoreClient(gcpClients *gcpclient.GcpClients) FilestoreClient {
	return NewFilestoreClientFromFilestoreClient(gcpClients.FilestoreWrapped())
}

func NewFilestoreClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FilestoreClient {
	return &filestoreClientImpl{
		filestoreClient: filestoreClient,
	}
}

// filestoreClientImpl implements FilestoreClient using the wrapped GCP Filestore interface.
type filestoreClientImpl struct {
	filestoreClient gcpclient.FilestoreClient
}

var _ FilestoreClient = &filestoreClientImpl{}

func (c *filestoreClientImpl) GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "instanceId", instanceId)

	req := &filestorepb.GetInstanceRequest{
		Name: GetFilestoreName(projectId, location, instanceId),
	}

	instance, err := c.filestoreClient.GetFilestoreInstance(ctx, req)

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Filestore instance not found")
			return nil, err
		}
		logger.Error(err, "Failed to get Filestore instance")
		return nil, err
	}

	return instance, nil
}

func (c *filestoreClientImpl) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.CreateInstanceRequest{
		Parent:     formatParentName(projectId, location),
		InstanceId: GetFilestoreInstanceId(instanceId),
		Instance:   instance,
	}

	op, err := c.filestoreClient.CreateFilestoreInstance(ctx, req)

	if err != nil {
		logger.Error(err, "Failed to create Filestore instance",
			"projectId", projectId,
			"location", location,
			"instanceId", instanceId)
		return "", err
	}

	return op.Name(), nil
}

func (c *filestoreClientImpl) UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	instance.Name = GetFilestoreName(projectId, location, instanceId)

	req := &filestorepb.UpdateInstanceRequest{
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
		Instance: instance,
	}

	op, err := c.filestoreClient.UpdateFilestoreInstance(ctx, req)

	if err != nil {
		logger.Error(err, "Failed to update Filestore instance",
			"projectId", projectId,
			"location", location,
			"instanceId", instanceId,
			"updateMask", updateMask)
		return "", err
	}

	return op.Name(), nil
}

func (c *filestoreClientImpl) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error) {
	logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "instanceId", instanceId)

	req := &filestorepb.DeleteInstanceRequest{
		Name: GetFilestoreName(projectId, location, instanceId),
	}

	op, err := c.filestoreClient.DeleteFilestoreInstance(ctx, req)

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target Filestore instance not found")
			return "", err
		}
		logger.Error(err, "Failed to delete Filestore instance")
		return "", err
	}

	return op.Name(), nil
}

func (c *filestoreClientImpl) GetOperation(ctx context.Context, operationName string) (bool, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &longrunningpb.GetOperationRequest{
		Name: operationName,
	}

	op, err := c.filestoreClient.GetFilestoreOperation(ctx, req)

	if err != nil {
		logger.Error(err, "Failed to get operation",
			"operationName", operationName)
		return false, err
	}

	return op.Done, nil
}

// Helper functions for constructing GCP resource names

func formatParentName(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}
