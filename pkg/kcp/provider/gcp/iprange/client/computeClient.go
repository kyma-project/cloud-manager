package client

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type ComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error)
	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error)
	DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error)
	GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error)
	GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error)
}

func NewComputeClient() client.ClientProvider[ComputeClient] {
	return client.NewCachedClientProvider(
		func(ctx context.Context, saJsonKeyPath string) (ComputeClient, error) {
			httpClient, err := client.GetCachedGcpClient(ctx, saJsonKeyPath)
			if err != nil {
				return nil, err
			}

			computeClient, err := compute.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, fmt.Errorf("error obtaining GCP Compute Client: [%w]", err)
			}
			return NewComputeClientForService(computeClient), nil
		},
	)
}

func NewComputeClientForService(svcCompute *compute.Service) ComputeClient {
	return &computeClient{svcCompute: svcCompute}
}

type computeClient struct {
	svcCompute *compute.Service
}

func (c *computeClient) GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcCompute.GlobalAddresses.Get(projectId, name).Do()
	if err != nil {
		logger.V(4).Info("GetIpRange", "err", err)
	}
	return out, err
}

func (c *computeClient) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcCompute.GlobalAddresses.Delete(projectId, name).Do()
	logger.V(4).Info("DeleteIpRange", "operation", operation, "err", err)
	return operation, err
}

func (c *computeClient) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	operation, err := c.svcCompute.GlobalAddresses.Insert(projectId, &compute.Address{
		Name:         name,
		Description:  description,
		Address:      address,
		PrefixLength: prefixLength,
		Network:      client.GetVPCPath(projectId, vpcName),
		AddressType:  string(client.AddressTypeInternal),
		Purpose:      string(client.IpRangePurposeVPCPeering),
	}).Do()
	logger.V(4).Info("CreatePscIpRange", "operation", operation, "err", err)
	return operation, err
}

func (c *computeClient) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	logger := composed.LoggerFromCtx(ctx)
	filter := client.GetNetworkFilter(projectId, vpc)
	out, err := c.svcCompute.GlobalAddresses.List(projectId).Filter(filter).Do()
	if err != nil {
		logger.Error(err, "ListGlobalAddresses", "projectId", projectId, "vpc", vpc)
		return nil, err
	}
	return out, nil
}

func (c *computeClient) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error) {
	logger := composed.LoggerFromCtx(ctx)
	out, err := c.svcCompute.GlobalOperations.Get(projectId, operationName).Do()
	if err != nil {
		logger.Error(err, "GetGlobalOperation", "projectId", projectId, "operationName", operationName)
		return nil, err
	}
	return out, nil
}
