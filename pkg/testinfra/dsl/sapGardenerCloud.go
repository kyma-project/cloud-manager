package dsl

import (
	"context"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/kyma-project/cloud-manager/pkg/common"
	sapmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/mock"
)

type SapGardenerInfra struct {
	VPC     *networks.Network
	Subnets []*subnets.Subnet
	Router  *routers.Router
}

func CreateSapGardenerResources(
	_ context.Context,
	sapMock sapmock.Server,
	shootNamespace, shootName string,
	nodesCidr string,
) (*SapGardenerInfra, error) {
	result := &SapGardenerInfra{}

	vpcName := common.GardenerVpcName(shootNamespace, shootName)

	result.VPC = sapMock.AddNetwork(uuid.NewString(), vpcName)
	result.Router = sapMock.AddRouter(uuid.NewString(), vpcName, "150.160.170.180")
	result.Subnets = []*subnets.Subnet{
		sapMock.AddSubnet(uuid.NewString(), result.VPC.ID, vpcName, nodesCidr),
	}

	return result, nil
}
