package client

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/protobuf/proto"
)

// ComputeClient embeds the wrapped gcpclient.GlobalAddressesClient and
// gcpclient.ComputeGlobalOperationsClient interfaces.
// Actions call the wrapped methods directly (e.g., GetGlobalAddress, DeleteGlobalAddress,
// ListGlobalAddresses, GetComputeGlobalOperation) by constructing protobuf requests inline.
// The CreatePscIpRange method is kept as a value-add method (complex Address proto construction
// with specific AddressType, Purpose, and network path building).
type ComputeClient interface {
	gcpclient.GlobalAddressesClient
	gcpclient.ComputeGlobalOperationsClient

	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error)
}

// NewComputeClientProvider returns a provider function that creates ComputeClient instances.
// Uses NEW pattern - accesses clients from GcpClients singleton.
func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
	return func(_ string) ComputeClient {
		return NewComputeClient(gcpClients)
	}
}

// NewComputeClient creates a new ComputeClient wrapping GcpClients.
func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
	return NewComputeClientFromWrapped(gcpClients.GlobalAddressesWrapped(), gcpClients.GlobalOperationsWrapped())
}

// NewComputeClientFromWrapped creates a ComputeClient from wrapped interfaces.
// Used by mock2 for test wiring.
func NewComputeClientFromWrapped(globalAddresses gcpclient.GlobalAddressesClient, globalOperations gcpclient.ComputeGlobalOperationsClient) ComputeClient {
	return &computeClient{
		GlobalAddressesClient:         globalAddresses,
		ComputeGlobalOperationsClient: globalOperations,
	}
}

type computeClient struct {
	gcpclient.GlobalAddressesClient
	gcpclient.ComputeGlobalOperationsClient
}

var _ ComputeClient = &computeClient{}

func (c *computeClient) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	req := &computepb.InsertGlobalAddressRequest{
		Project: projectId,
		AddressResource: &computepb.Address{
			Name:         proto.String(name),
			Description:  proto.String(description),
			Address:      proto.String(address),
			PrefixLength: proto.Int32(int32(prefixLength)),
			Network:      proto.String(gcpclient.GetVPCPath(projectId, vpcName)),
			AddressType:  proto.String(computepb.Address_INTERNAL.String()),
			Purpose:      proto.String(computepb.Address_VPC_PEERING.String()),
		},
	}

	op, err := c.InsertGlobalAddress(ctx, req)

	var operationName string
	if op != nil {
		operationName = op.Name()
	}

	return operationName, err
}
