package dsl

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/common"
	azuremock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/mock"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

type AzureGardenerInfra struct {
	VNet              *armnetwork.VirtualNetwork
	PublicIpAddresses []*armnetwork.PublicIPAddress
	NatGateways       []*armnetwork.NatGateway
	Subnets           []*armnetwork.Subnet
}

func CreateAzureGardenerResources(
	ctx context.Context,
	azureMock azuremock.TenantSubscription,
	shootNamespace, shootName string,
	vnetCidr, nodesCidr string,
	location string,
) (*AzureGardenerInfra, error) {
	result := &AzureGardenerInfra{}

	resourceGroupName := common.GardenerVpcName(shootNamespace, shootName)

	wholeRange, err := cidr.Parse(vnetCidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIDR '%s': %v", vnetCidr, err)
	}

	err = azureMock.CreateNetwork(ctx, resourceGroupName, resourceGroupName, location, vnetCidr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating vnet: %w", err)
	}

	vnet, err := azureMock.GetNetwork(ctx, resourceGroupName, resourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("error getting vnet: %w", err)
	}
	result.VNet = vnet

	numberOfSubnets := 4
	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, numberOfSubnets)
	if err != nil {
		return nil, fmt.Errorf("error splitting vnet cidr %s: %w", vnetCidr, err)
	}

	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("%s-%d", resourceGroupName, i)
		zone := fmt.Sprintf("%d", i)
		err = azureMock.CreatePublicIpAddress(ctx, resourceGroupName, name, location, zone)
		if err != nil {
			return nil, fmt.Errorf("error creating public ip address in zone %d: %w", i, err)
		}
		ipAddressId := azureutil.NewPublicIpAddressResourceId(azureMock.SubscriptionId(), resourceGroupName, name)

		ip, err := azureMock.GetPublicIpAddress(ctx, resourceGroupName, name)
		if err != nil {
			return nil, fmt.Errorf("error getting public ip address in zone %d: %w", i, err)
		}
		result.PublicIpAddresses = append(result.PublicIpAddresses, ip)

		err = azureMock.CreateNatGateway(ctx, resourceGroupName, name, location, fmt.Sprintf("%d", i), nil, []string{ipAddressId.String()})
		if err != nil {
			return nil, fmt.Errorf("error creating nat gateway in zone %d: %w", i, err)
		}
		netGatewayId := azureutil.NewNatGatewayResourceId(azureMock.SubscriptionId(), resourceGroupName, name)

		nat, err := azureMock.GetNatGateway(ctx, resourceGroupName, name)
		if err != nil {
			return nil, fmt.Errorf("error getting nat gateway in zone %d: %w", i, err)
		}
		result.NatGateways = append(result.NatGateways, nat)

		err = azureMock.CreateSubnet(ctx, resourceGroupName, resourceGroupName, name, subnetRanges[i].CIDR().String(), "x", netGatewayId.String())
		if err != nil {
			return nil, fmt.Errorf("error creating subnet in zone %d: %w", i, err)
		}

		subnet, err := azureMock.GetSubnet(ctx, resourceGroupName, resourceGroupName, name)
		if err != nil {
			return nil, fmt.Errorf("error getting subnet in zone %d: %w", i, err)
		}
		result.Subnets = append(result.Subnets, subnet)
	}

	return result, nil
}
