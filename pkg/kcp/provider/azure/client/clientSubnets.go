package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type SubnetsClient interface {
	GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error)

	CreateSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error)

	DeleteSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (Poller[armnetwork.SubnetsClientDeleteResponse], error)
}

func NewSubnetsClient(svc *armnetwork.SubnetsClient) SubnetsClient {
	return &subnetsClient{svc: svc}
}

// helper functions ===================================================================

func NewSubnet(addressPrefix string, securityGroupId string, natGatewayId string) armnetwork.Subnet {
	subnet := armnetwork.Subnet{
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: ptr.To(addressPrefix),
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
	return subnet
}

// subnetsClient impl ===================================================================

var _ SubnetsClient = &subnetsClient{}

type subnetsClient struct {
	svc *armnetwork.SubnetsClient
}

func (c *subnetsClient) CreateSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	return c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, subnetName, subnetParameters, options)
}

func (c *subnetsClient) GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

func (c *subnetsClient) DeleteSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	return c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, subnetName, options)
}
