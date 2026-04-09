package client

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/protobuf/proto"
)

// ComputeClient defines business operations for IpRange Compute API interactions.
// Uses NEW pattern with cloud.google.com/go/compute/apiv1 (Cloud Client Library).
type ComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error)
	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) // returns operation name
	DeleteIpRange(ctx context.Context, projectId, name string) (string, error)                                                       // returns operation name
	GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error)
	GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error)
	WaitGlobalOperation(ctx context.Context, projectId, operationName string) error
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
		globalAddressesClient:  globalAddresses,
		globalOperationsClient: globalOperations,
	}
}

type computeClient struct {
	globalAddressesClient  gcpclient.GlobalAddressesClient
	globalOperationsClient gcpclient.ComputeGlobalOperationsClient
}

func (c *computeClient) GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &computepb.GetGlobalAddressRequest{
		Project: projectId,
		Address: name,
	}

	out, err := c.globalAddressesClient.GetGlobalAddress(ctx, req)
	if err != nil {
		logger.Info("GetIpRange", "err", err)
	}
	return out, err
}

func (c *computeClient) DeleteIpRange(ctx context.Context, projectId, name string) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &computepb.DeleteGlobalAddressRequest{
		Project: projectId,
		Address: name,
	}

	op, err := c.globalAddressesClient.DeleteGlobalAddress(ctx, req)

	var operationName string
	if op != nil {
		operationName = op.Name()
	}

	logger.Info("DeleteIpRange", "operation", operationName, "err", err)
	return operationName, err
}

func (c *computeClient) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	logger := composed.LoggerFromCtx(ctx)

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

	op, err := c.globalAddressesClient.InsertGlobalAddress(ctx, req)

	var operationName string
	if op != nil {
		operationName = op.Name()
	}

	logger.Info("CreatePscIpRange", "operation", operationName, "err", err)
	return operationName, err
}

func (c *computeClient) ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error) {
	logger := composed.LoggerFromCtx(ctx)

	filter := gcpclient.GetNetworkFilter(projectId, vpc)

	req := &computepb.ListGlobalAddressesRequest{
		Project: projectId,
		Filter:  proto.String(filter),
	}

	it := c.globalAddressesClient.ListGlobalAddresses(ctx, req)

	var addresses []*computepb.Address
	for addr, err := range it.All() {
		if err != nil {
			logger.Error(err, "ListGlobalAddresses", "projectId", projectId, "vpc", vpc)
			return nil, err
		}
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func (c *computeClient) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)

	req := &computepb.GetGlobalOperationRequest{
		Project:   projectId,
		Operation: operationName,
	}

	out, err := c.globalOperationsClient.GetComputeGlobalOperation(ctx, req)
	if err != nil {
		logger.Error(err, "GetGlobalOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return out, nil
}

func (c *computeClient) WaitGlobalOperation(ctx context.Context, projectId, operationName string) error {
	// WaitGlobalOperation is not available on the wrapped interface.
	// This method is retained for interface compatibility but is not called in production code.
	// Use GetGlobalOperation and poll instead.
	logger := composed.LoggerFromCtx(ctx)
	logger.Info("WaitGlobalOperation called - polling via GetGlobalOperation", "projectId", projectId, "operationName", operationName)
	_, err := c.GetGlobalOperation(ctx, projectId, operationName)
	return err
}
