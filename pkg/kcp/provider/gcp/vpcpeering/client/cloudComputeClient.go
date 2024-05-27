package client

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	"google.golang.org/api/option"
)

/*
required GCP permissions
=========================
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Removes the VPC peering connection
  compute.networks.removePeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/removePeering
  ** Gets the network (VPCs) in order to retrieve the peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get
*/

func NewClientProvider() cloudclient.SkrClientProvider[Client] {
	return func(ctx context.Context, saJsonKeyPath string) (Client, error) {
		c, err := compute.NewNetworksRESTClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
		if err != nil {
			return nil, err
		}
		//defer c.Close()
		return &networkClient{cnc: c}, nil
	}
}

type networkClient struct {
	cnc *compute.NetworksClient
}

type Client interface {
	CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (pb.NetworkPeering, error)
}

func (c *networkClient) CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (pb.NetworkPeering, error) {
	networkPeering := pb.NetworkPeering{}

	kymaNetwork := getFullNetworkUrl(*kymaProject, *kymaVpc)
	remoteNetwork := getFullNetworkUrl(*remoteProject, *remoteVpc)

	//peering from kyma to remote vpc
	peeringRequestFromKyma := &pb.AddPeeringNetworkRequest{
		Network: *kymaVpc,
		Project: *kymaProject,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:                 name,
				Network:              &remoteNetwork,
				ImportCustomRoutes:   importCustomRoutes,
				ExchangeSubnetRoutes: to.Ptr(true),
			},
		},
	}

	peeringOperationFromKyma, err := c.cnc.AddPeering(ctx, peeringRequestFromKyma)
	if err != nil {
		return networkPeering, err
	}
	err = peeringOperationFromKyma.Wait(ctx)
	if err != nil {
		return networkPeering, err
	}

	net, err := c.cnc.Get(ctx, &pb.GetNetworkRequest{Network: *kymaVpc, Project: *kymaProject})
	nps := net.GetPeerings()
	for _, np := range nps {
		if *np.Network == remoteNetwork {
			networkPeering = *np
			break
		}
	}

	//peering from remote vpc to kyma
	//by default exportCustomRoutes is false but if the remote vpc wants kyma to import custom routes, the peering needs to export them :)
	exportCustomRoutes := false
	if *importCustomRoutes {
		exportCustomRoutes = true
	}
	peeringRequestFromRemote := &pb.AddPeeringNetworkRequest{
		Network: *remoteVpc,
		Project: *remoteProject,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:                 name,
				Network:              &kymaNetwork,
				ExportCustomRoutes:   &exportCustomRoutes,
				ExchangeSubnetRoutes: Pointer(true),
			},
		},
	}

	peeringOperationFromRemote, err := c.cnc.AddPeering(ctx, peeringRequestFromRemote)
	if err != nil {
		return networkPeering, err
	}
	err = peeringOperationFromRemote.Wait(ctx)
	if err != nil {
		return networkPeering, err
	}
	return networkPeering, nil
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func Pointer[T any](d T) *T {
	return &d
}
