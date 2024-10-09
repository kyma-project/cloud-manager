package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

type PrivateDnsZoneGroupClient interface {
	GetPrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupGroupName string) (*armnetwork.PrivateDNSZoneGroup, error)
	CreatePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string, parameters armnetwork.PrivateDNSZoneGroup) error
}

func NewPrivateDnsZoneGroupClient(svc *armnetwork.PrivateDNSZoneGroupsClient) PrivateDnsZoneGroupClient {
	return &privateDnsZoneGroupClient{svc: svc}
}

var _ PrivateDnsZoneGroupClient = &privateDnsZoneGroupClient{}

type privateDnsZoneGroupClient struct {
	svc *armnetwork.PrivateDNSZoneGroupsClient
}

func (c *privateDnsZoneGroupClient) GetPrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) (*armnetwork.PrivateDNSZoneGroup, error) {
	privateDnsZoneGroupClientGetResponse, err := c.svc.Get(
		ctx,
		resourceGroupName,
		privateEndPointName,
		privateDnsZoneGroupName,
		nil)
	if err != nil {
		return nil, err
	}
	return &privateDnsZoneGroupClientGetResponse.PrivateDNSZoneGroup, nil
}

func (c *privateDnsZoneGroupClient) CreatePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string, parameters armnetwork.PrivateDNSZoneGroup) error {
	_, err := c.svc.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateEndPointName,
		privateDnsZoneGroupName,
		parameters,
		nil)
	if err != nil {
		return err
	}
	return nil
}
