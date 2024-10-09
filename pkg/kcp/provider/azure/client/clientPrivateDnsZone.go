package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

type PrivateDnsZoneClient interface {
	CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error
	GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) (*armprivatedns.PrivateZone, error)
	DeletePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) error
}

func NewPrivateDnsZoneClient(svc *armprivatedns.PrivateZonesClient) PrivateDnsZoneClient {
	return &privateDnsZoneClient{svc: svc}
}

var _ PrivateDnsZoneClient = &privateDnsZoneClient{}

type privateDnsZoneClient struct {
	svc *armprivatedns.PrivateZonesClient
}

func (c *privateDnsZoneClient) CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error {
	_, err := c.svc.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateDnsZoneName,
		parameters,
		nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *privateDnsZoneClient) GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) (*armprivatedns.PrivateZone, error) {
	privateDnsZoneClientGetResponse, err := c.svc.Get(
		ctx,
		resourceGroupName,
		privateDnsZoneName,
		nil)
	if err != nil {
		return nil, err
	}
	return &privateDnsZoneClientGetResponse.PrivateZone, nil
}

func (c *privateDnsZoneClient) DeletePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) error {
	_, err := c.svc.BeginDelete(
		ctx,
		resourceGroupName,
		privateDnsZoneName,
		nil)
	if err != nil {
		return err
	}
	return nil
}
