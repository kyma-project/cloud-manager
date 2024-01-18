package client

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"net/http"

	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type ComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error)
	CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error)
	DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error)
	GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error)
}

func NewComputeClient() gcpclient.ClientProvider[ComputeClient] {
	return gcpclient.NewCachedClientProvider(
		func(ctx context.Context, httpClient *http.Client) (ComputeClient, error) {
			client, err := compute.NewService(ctx, option.WithHTTPClient(httpClient))
			if err != nil {
				return nil, err
			}
			return newComputeClient(client), nil
		},
	)
}

func newComputeClient(svcCompute *compute.Service) ComputeClient {
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
		Network:      gcpclient.GetVPCPath(projectId, vpcName),
		AddressType:  string(gcpclient.AddressTypeInternal),
		Purpose:      string(gcpclient.IpRangePurposeVPCPeering),
	}).Do()
	logger.V(4).Info("CreatePscIpRange", "operation", operation, "err", err)
	return operation, err
}

func (c *computeClient) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	logger := composed.LoggerFromCtx(ctx)
	filter := fmt.Sprintf(gcpclient.NetworkFilter, projectId, vpc)
	out, err := c.svcCompute.GlobalAddresses.List(projectId).Filter(filter).Do()
	if err != nil {
		logger.Error(err, "ListGlobalAddresses", "projectId", projectId, "vpc", vpc)
		return nil, err
	}
	return out, nil
}
