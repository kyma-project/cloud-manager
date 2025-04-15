package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

type PublicIPAddressesClient interface {
	GetPublicIpAddress(ctx context.Context, resourceGroupName, publicIPAddressName string) (*armnetwork.PublicIPAddress, error)
}

func NewPublicIPAddressesClient(svc *armnetwork.PublicIPAddressesClient) PublicIPAddressesClient {
	return &publicIPAddressesClient{svc: svc}
}

var _ PublicIPAddressesClient = (*publicIPAddressesClient)(nil)

type publicIPAddressesClient struct {
	svc *armnetwork.PublicIPAddressesClient
}

func (c *publicIPAddressesClient) GetPublicIpAddress(ctx context.Context, resourceGroupName, publicIPAddressName string) (*armnetwork.PublicIPAddress, error) {
	//TODO implement me
	panic("implement me")
}
