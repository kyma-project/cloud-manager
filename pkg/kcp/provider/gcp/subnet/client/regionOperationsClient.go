package client

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type GetRegionOperationRequest struct {
	ProjectId string
	Region    string
	Name      string
}

type RegionOperationsClient interface {
	GetRegionOperation(ctx context.Context, request GetRegionOperationRequest) (*computepb.Operation, error)
}

func NewRegionOperationsClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[RegionOperationsClient] {
	return func(_ string) RegionOperationsClient {
		return NewRegionOperationsClient(gcpClients)
	}
}

func NewRegionOperationsClient(gcpClients *gcpclient.GcpClients) RegionOperationsClient {
	return NewRegionOperationsClientFromWrapped(gcpClients.RegionOperationsWrapped())
}

func NewRegionOperationsClientFromWrapped(regionalOpsClient gcpclient.ComputeRegionalOperationsClient) RegionOperationsClient {
	return &regionalOperationsClient{regionalOpsClient: regionalOpsClient}
}

type regionalOperationsClient struct {
	regionalOpsClient gcpclient.ComputeRegionalOperationsClient
}

func (c *regionalOperationsClient) GetRegionOperation(ctx context.Context, request GetRegionOperationRequest) (*computepb.Operation, error) {
	op, err := c.regionalOpsClient.GetComputeRegionalOperation(ctx, &computepb.GetRegionOperationRequest{
		Project:   request.ProjectId,
		Region:    request.Region,
		Operation: request.Name,
	})
	if err != nil {
		return nil, err
	}
	return op, nil
}
