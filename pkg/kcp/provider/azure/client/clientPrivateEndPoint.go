package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

type PrivateEndPointsClient interface {
	GetPrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error)
	CreatePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string, parameters armnetwork.PrivateEndpoint) error
	DeletePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) error
}

func NewPrivateEndPointClient(svc *armnetwork.PrivateEndpointsClient) PrivateEndPointsClient {
	return &privateEndPointClient{svc: svc}
}

var _ PrivateEndPointsClient = &privateEndPointClient{}

type privateEndPointClient struct {
	svc *armnetwork.PrivateEndpointsClient
}

func (c *privateEndPointClient) GetPrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error) {
	privateEndpointsClientGetResponse, err := c.svc.Get(
		ctx,
		resourceGroupName,
		privateEndPointName,
		nil)
	if err != nil {
		return nil, err
	}
	return &privateEndpointsClientGetResponse.PrivateEndpoint, nil
}

func (c *privateEndPointClient) CreatePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string, parameters armnetwork.PrivateEndpoint) error {
	_, err := c.svc.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateEndPointName,
		parameters,
		nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *privateEndPointClient) DeletePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) error {
	_, err := c.svc.BeginDelete(
		ctx,
		resourceGroupName,
		privateEndPointName,
		nil)
	if err != nil {
		return err
	}
	return nil
}
