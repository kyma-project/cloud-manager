package client

import (
	"context"
	"fmt"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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
	return func() FilestoreClient {
		return NewFilestoreClient(gcpClients)
	}
}

// NewFilestoreClient creates a new FilestoreClient wrapping GcpClients.
func NewFilestoreClient(gcpClients *gcpclient.GcpClients) FilestoreClient {
	return &filestoreClient{
		cloudFilestoreManager: gcpClients.Filestore,
	}
}

// filestoreClient implements FilestoreClient using the modern GCP Filestore API.
type filestoreClient struct {
	cloudFilestoreManager *filestore.CloudFilestoreManagerClient
}

var _ FilestoreClient = &filestoreClient{}

func (c *filestoreClient) GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.GetInstanceRequest{
		Name: formatInstanceName(projectId, location, instanceId),
	}

	instance, err := c.cloudFilestoreManager.GetInstance(ctx, req)
	gcpclient.IncrementCallCounter("Filestore", "GetInstance", location, err)

	if err != nil {
		logger.Error(err, "Failed to get Filestore instance",
			"projectId", projectId,
			"location", location,
			"instanceId", instanceId)
		return nil, err
	}

	return instance, nil
}

func (c *filestoreClient) CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.CreateInstanceRequest{
		Parent:     formatParentName(projectId, location),
		InstanceId: instanceId,
		Instance:   instance,
	}

	op, err := c.cloudFilestoreManager.CreateInstance(ctx, req)
	gcpclient.IncrementCallCounter("Filestore", "CreateInstance", location, err)

	if err != nil {
		logger.Error(err, "Failed to create Filestore instance",
			"projectId", projectId,
			"location", location,
			"instanceId", instanceId)
		return "", err
	}

	return op.Name(), nil
}

func (c *filestoreClient) UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	// Set the Name field - required for UpdateInstance API
	instance.Name = formatInstanceName(projectId, location, instanceId)

	req := &filestorepb.UpdateInstanceRequest{
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updateMask,
		},
		Instance: instance,
	}

	op, err := c.cloudFilestoreManager.UpdateInstance(ctx, req)
	gcpclient.IncrementCallCounter("Filestore", "UpdateInstance", location, err)

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

func (c *filestoreClient) DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &filestorepb.DeleteInstanceRequest{
		Name: formatInstanceName(projectId, location, instanceId),
	}

	op, err := c.cloudFilestoreManager.DeleteInstance(ctx, req)
	gcpclient.IncrementCallCounter("Filestore", "DeleteInstance", location, err)

	if err != nil {
		logger.Error(err, "Failed to delete Filestore instance",
			"projectId", projectId,
			"location", location,
			"instanceId", instanceId)
		return "", err
	}

	return op.Name(), nil
}

func (c *filestoreClient) GetOperation(ctx context.Context, operationName string) (bool, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &longrunningpb.GetOperationRequest{
		Name: operationName,
	}

	op, err := c.cloudFilestoreManager.LROClient.GetOperation(ctx, req)
	gcpclient.IncrementCallCounter("Filestore", "GetOperation", "", err)

	if err != nil {
		logger.Error(err, "Failed to get operation",
			"operationName", operationName)
		return false, err
	}

	return op.Done, nil
}

// Helper functions for constructing GCP resource names

func formatInstanceName(projectId, location, instanceId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, instanceId)
}

func formatParentName(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}
