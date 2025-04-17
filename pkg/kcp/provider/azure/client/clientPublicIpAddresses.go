package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

type PublicIPAddressesClient interface {
	CreatePublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName, location, zone string) error
	GetPublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName string) (*armnetwork.PublicIPAddress, error)
}

func NewPublicIPAddressesClient(svc *armnetwork.PublicIPAddressesClient) PublicIPAddressesClient {
	return &publicIPAddressesClient{svc: svc}
}

var _ PublicIPAddressesClient = (*publicIPAddressesClient)(nil)

type publicIPAddressesClient struct {
	svc *armnetwork.PublicIPAddressesClient
}

func (c *publicIPAddressesClient) CreatePublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName, location, zone string) error {
	return errors.New("not implemented")
}

func (c *publicIPAddressesClient) GetPublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName string) (*armnetwork.PublicIPAddress, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, publicIpAddressName, &armnetwork.PublicIPAddressesClientGetOptions{})
	if err != nil {
		return nil, err
	}
	return &resp.PublicIPAddress, nil
}
