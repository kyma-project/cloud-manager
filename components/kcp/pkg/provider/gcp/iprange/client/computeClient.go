package client

import (
	"context"
	"fmt"

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
		func(ctx context.Context, saJsonKeyPath string) (ComputeClient, error) {
			client, err := compute.NewService(ctx, option.WithCredentialsFile(saJsonKeyPath))
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
	return c.svcCompute.GlobalAddresses.Get(projectId, name).Do()
}

func (c *computeClient) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	return c.svcCompute.GlobalAddresses.Delete(projectId, name).Do()
}

func (c *computeClient) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	return c.svcCompute.GlobalAddresses.Insert(projectId, &compute.Address{
		Name:         name,
		Description:  description,
		Address:      address,
		PrefixLength: prefixLength,
		Network:      gcpclient.GetVPCPath(projectId, vpcName),
		AddressType:  string(gcpclient.AddressTypeInternal),
		Purpose:      string(gcpclient.IpRangePurposeVPCPeering),
	}).Do()
}

func (c *computeClient) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	filter := fmt.Sprintf(gcpclient.NetworkFilter, projectId, vpc)
	out, err := c.svcCompute.GlobalAddresses.List(projectId).Filter(filter).Do()
	if err != nil {
		return nil, err
	}
	return out, nil
}
