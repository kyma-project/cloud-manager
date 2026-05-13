package dsl

import (
	"context"
	"fmt"
	"math/rand"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock"
	"k8s.io/utils/ptr"
)

type GcpGardenerCloudInfra struct {
	// the mock does not have generic vpc and subnet funcs like SDK does, so won't do it for now
	//Vpc *computepb.Network
	//Subnet  []*computepb.Subnetwork
	Router  []*computepb.Router
	Address []*computepb.Address
}

func CreateGcpGardenerResources(
	_ context.Context,
	gcpMock gcpmock.Server,
	shootNamespace, shootName string,
	nodesCidr string,
	gcpProject string,
	region string,
) (*GcpGardenerCloudInfra, error) {
	result := &GcpGardenerCloudInfra{}

	vpcName := common.GardenerVpcName(shootNamespace, shootName)

	router, err := gcpMock.InsertVpcRouter(gcpProject, region, &computepb.Router{
		Name:    new(vpcName + "-cloud-router"),
		Network: new(vpcName),
	})
	if err != nil {
		return nil, err
	}
	result.Router = append(result.Router, router)

	address, err := gcpMock.InsertAddress(gcpProject, region, &computepb.Address{
		Name:        new(fmt.Sprintf("nat-auto-ip-%d", rand.Intn(10000000))),
		Purpose:     new("NAT_AUTO"),
		NetworkTier: new("PREMIUM"),
		Users:       []string{ptr.Deref(router.SelfLink, "")},
	})
	if err != nil {
		return nil, err
	}
	result.Address = append(result.Address, address)

	return result, nil
}
