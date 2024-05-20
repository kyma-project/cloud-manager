package client

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/option"
)

func NewNetworksClientProvider() client.ClientProvider[Client] {
	c, err := compute.NewNetworksRESTClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c
}

type networkClient struct {
	cnc *compute.NetworksClient
}

type Client interface {
	DescribeVpcs(ctx context.Context, project string) ([]pb.Network, error)
	CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, importCustomRoutes bool, kymaProject string, kymaVpc string) (pb.NetworkPeering, error)
	DescribeVpcPeeringConnections(ctx context.Context, project string) ([]pb.NetworkPeering, error)
}

func (c *networkClient) DescribeVpcs(ctx context.Context, project string) ([]pb.Network, error) {
	networkIterator := c.cnc.List(ctx, &pb.ListNetworksRequest{Project: project})
	var networks []pb.Network
	for {
		network, err := networkIterator.Next()
		if err != nil {
			return nil, err
		}
		networks = append(networks, *network)
	}
	return networks, nil
}

func (c *networkClient) CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, importCustomRoutes bool, kymaProject string, kymaVpc string) (pb.NetworkPeering, error) {
	networkPeering := pb.NetworkPeering{}

	//peering from kyma to remote vpc
	peeringRequestFromKyma := &pb.AddPeeringNetworkRequest{
		Network: *remoteVpc,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:               name,
				Network:            remoteVpc,
				ImportCustomRoutes: &importCustomRoutes,
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

	net, err := c.cnc.Get(ctx, &pb.GetNetworkRequest{Network: *remoteVpc})
	nps := net.GetPeerings()
	for _, np := range nps {
		if np.Network == remoteVpc {
			networkPeering = *np
		}
	}

	//peering from remote vpc to kyma
	kymaNetwork := "projects/" + kymaProject + "/global/networks/" + kymaVpc
	//by default exportCustomRoutes is false but if the remote vpc wants kyma to import custom routes, the peering needs to export them :)
	exportCustomRoutes := false
	if importCustomRoutes {
		exportCustomRoutes = true
	}
	peeringRequestFromRemote := &pb.AddPeeringNetworkRequest{
		Network: kymaNetwork,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:               name,
				Network:            &kymaNetwork,
				ExportCustomRoutes: &exportCustomRoutes,
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

//func (c *networkClient) DescribeVpcPeeringConnections(ctx context.Context, vpc string) ([]pb.NetworkPeering, error) {
//filter := "network eq " + vpc //get all peering connections
//networkIterator := c.cnc.List(ctx, &pb.ListNetworksRequest{Filter: &filter})
//var peerings []*pb.NetworkPeering
//for {
//	network, err := networkIterator.Next()
//	if err != nil {
//		return nil, err
//	}
//	append(peerings, network.Peerings)
//}

//	return nil, nil
//}
