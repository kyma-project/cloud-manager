package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

type NatGatewayClient interface {
	CreateNatGateway(ctx context.Context, resourceGroupName, natGatewayName, location, zone string, subnetIds, publicIpAddressIds []string) error
	GetNatGateway(ctx context.Context, resourceGroupName, natGatewayName string) (*armnetwork.NatGateway, error)
}

func NewNatGatewayClient(svc *armnetwork.NatGatewaysClient) NatGatewayClient {
	return &natGatewayClient{svc: svc}
}

var _ NatGatewayClient = (*natGatewayClient)(nil)

type natGatewayClient struct {
	svc *armnetwork.NatGatewaysClient
}

func (c *natGatewayClient) CreateNatGateway(ctx context.Context, resourceGroupName, natGatewayName, location, zone string, subnetIds, publicIpAddressIds []string) error {
	return errors.New("not implemented")
}

func (c *natGatewayClient) GetNatGateway(ctx context.Context, resourceGroupName, natGatewayName string) (*armnetwork.NatGateway, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, natGatewayName, &armnetwork.NatGatewaysClientGetOptions{})
	if err != nil {
		return nil, err
	}

	return &resp.NatGateway, nil
}
