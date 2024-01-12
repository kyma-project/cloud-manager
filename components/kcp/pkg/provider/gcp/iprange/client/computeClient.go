package client

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type ComputeClient interface {
	ListGlobalAddresses(ctx context.Context, projectId string) (*compute.AddressList, error)
	CreatePscIpRange(ctx context.Context, projectId, name, description, address string, prefixLength int64) (*compute.Operation, error)
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

// CreatePscIpRange implements ComputeClient.
func (c *computeClient) CreatePscIpRange(ctx context.Context, projectId, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	return c.svcCompute.GlobalAddresses.Insert(projectId, &compute.Address{
		Name:        name,
		Description: description,
		Address:     address,
		PrefixLength: prefixLength,
		NetworkTier: string(client.NetworkTierPremium),
		AddressType: string(client.AddressTypeInternal),
		Purpose:    string(client.IpRangePurposeVPCPeering),
	}).Do()
}

func (c *computeClient) ListGlobalAddresses(ctx context.Context, projectId string) (*compute.AddressList, error) {
	out, err := c.svcCompute.GlobalAddresses.List(projectId).Do()
	if err != nil {
		return nil, err
	}
	return out, nil
}
