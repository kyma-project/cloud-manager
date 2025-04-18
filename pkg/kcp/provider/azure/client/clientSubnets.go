package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type SubnetsClient interface {
	GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error)
	CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, addressPrefix, securityGroupId, natGatewayId string) error
	DeleteSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) error
}

func NewSubnetsClient(svc *armnetwork.SubnetsClient) SubnetsClient {
	return &subnetsClient{svc: svc}
}

var _ SubnetsClient = &subnetsClient{}

type subnetsClient struct {
	svc *armnetwork.SubnetsClient
}

func (c *subnetsClient) CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, addressPrefix, securityGroupId, natGatewayId string) error {
	subnet := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix:         ptr.To(addressPrefix),
			DefaultOutboundAccess: ptr.To(false),
		},
	}
	if securityGroupId != "" {
		subnet.Properties.NetworkSecurityGroup = &armnetwork.SecurityGroup{
			ID: ptr.To(securityGroupId),
		}
	}
	if natGatewayId != "" {
		subnet.Properties.NatGateway = &armnetwork.SubResource{
			ID: ptr.To(natGatewayId),
		}
	}
	_, err := c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, subnet, nil)
	return err
}

func (c *subnetsClient) GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

func (c *subnetsClient) DeleteSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)
	return err
}
