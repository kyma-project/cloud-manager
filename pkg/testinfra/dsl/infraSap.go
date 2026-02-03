package dsl

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/kyma-project/cloud-manager/pkg/common"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/mock"
	sapvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork"
)

// SapInfraInitializeExternalNetworks ensures the external network and its subnet required for floating IPs exist.
// OpenStack is special and an exception in providers with this, since in the project there must be an external network
// that represents the Internet
func SapInfraInitializeExternalNetworks(ctx context.Context, sapMock sapmock.Project) error {
	arrNets, err := sapMock.ListExternalNetworks(ctx, networks.ListOpts{
		Name: sapconfig.SapConfig.FloatingPoolNetwork,
	})
	if err != nil {
		return fmt.Errorf("error listing external networks: %v", err)
	}

	var net *networks.Network
	if len(arrNets) > 0 {
		net = &arrNets[0]
	} else {
		n, err := sapMock.CreateExternalNetwork(ctx, networks.CreateOpts{
			Name: sapconfig.SapConfig.FloatingPoolNetwork,
		})
		if err != nil {
			return fmt.Errorf("error creating external network: %v", err)
		}
		net = n
	}

	arrSubnets, err := sapMock.ListSubnets(ctx, subnets.ListOpts{
		NetworkID: net.ID,
	})
	if err != nil {
		return fmt.Errorf("error listing external network subnets: %v", err)
	}

	found := false
	externalSubnetRegex, err := regexp.Compile(sapconfig.SapConfig.FloatingPoolSubnet)
	if err != nil {
		return fmt.Errorf("failed to parse floating pool subnet regexp: %w", err)
	}
	for _, s := range arrSubnets {
		if externalSubnetRegex.MatchString(s.Name) {
			found = true
			break
		}
	}
	if found {
		return nil
	}

	subnetName := strings.TrimSuffix(sapconfig.SapConfig.FloatingPoolSubnet, "*")
	subnetName = strings.TrimSuffix(subnetName, ".*")

	_, err = sapMock.CreateSubnet(ctx, subnets.CreateOpts{
		NetworkID: net.ID,
		CIDR:      "30.40.0.0/16",
		Name:      subnetName,
		IPVersion: 4,
	})
	if err != nil {
		return fmt.Errorf("error creating subnet: %v", err)
	}

	return nil
}

func SapInfraGardenerCreateResourcesAllWithNetwork(
	ctx context.Context,
	sapMock sapmock.Project,
	shootNamespace, shootName string,
	nodesCidr string,
) (*SapGardenerInfra, error) {
	out, err := sapvpcnetwork.CreateInfra(
		ctx,
		sapvpcnetwork.WithName(common.GardenerVpcName(shootNamespace, shootName)),
		sapvpcnetwork.WithClient(sapMock),
		sapvpcnetwork.WithCidrBlocks([]string{nodesCidr}),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating VPC network: %v", err)
	}
	return SapInfraGardenerCreateResourcesInKymaNetwork(ctx, sapMock, out, shootNamespace, shootName, nodesCidr)
}

func SapInfraGardenerCreateResourcesInKymaNetwork(
	ctx context.Context,
	sapMock sapmock.Project,
	out *sapvpcnetwork.CreateInfraOutput,
	shootNamespace, shootName string,
	nodesCidr string,
) (*SapGardenerInfra, error) {
	subnetName := common.GardenerVpcName(shootNamespace, shootName)
	s, err := sapMock.CreateSubnet(ctx, subnets.CreateOpts{
		NetworkID: out.InternalNetwork.ID,
		CIDR:      nodesCidr,
		Name:      subnetName,
		IPVersion: 4,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating subnet: %v", err)
	}
	return &SapGardenerInfra{
		VPC: out.InternalNetwork,
		Subnets: []*subnets.Subnet{
			s,
		},
		Router: out.Router,
	}, nil
}
