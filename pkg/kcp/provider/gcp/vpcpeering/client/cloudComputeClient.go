package client

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	"google.golang.org/api/option"
	"k8s.io/utils/ptr"
	"strings"
)

/*
required GCP permissions
=========================
  - Both Sides - The service account used to create the VPC peering connection needs the following permissions:
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Removes the VPC peering connection
  compute.networks.removePeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/removePeering
  ** Gets the network (VPCs) in order to retrieve the peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get

  - Remote Side - The service account used to create the VPC peering connection needs the additional permissions:
  ** Fetches the remote network tags
  compute.networks.ListEffectiveTags => https://cloud.google.com/resource-manager/reference/rest/v3/tagKeys/get
*/

func NewClientProvider() cloudclient.ClientProvider[Client] {
	return func(ctx context.Context, saJsonKeyPath string) (Client, error) {
		c, err := compute.NewNetworksRESTClient(ctx, option.WithCredentialsFile(saJsonKeyPath))
		if err != nil {
			return nil, err
		}
		return &networkClient{cnc: c}, nil
	}
}

type networkClient struct {
	cnc *compute.NetworksClient
}

type Client interface {
	CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (*pb.NetworkPeering, error)
}

func (c *networkClient) CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (*pb.NetworkPeering, error) {

	kymaNetwork := getFullNetworkUrl(*kymaProject, *kymaVpc)
	remoteNetwork := getFullNetworkUrl(*remoteProject, *remoteVpc)
	defer c.cnc.Close()

	//NetworkPeering will only be created if the remote vpc has a tag with the kyma shoot name
	remoteNetworkInfo, err := c.cnc.Get(ctx, &pb.GetNetworkRequest{Network: *remoteVpc, Project: *remoteProject})
	if err != nil {
		return nil, err
	}

	isRemoteNetworkTagged, err := c.CheckRemoteNetworkTags(ctx, remoteNetworkInfo, *kymaVpc)

	if isRemoteNetworkTagged == false || (err != nil && err.Error() == "no more items in iterator") {
		return nil, fmt.Errorf("remote network " + *remoteVpc + " is not tagged with the kyma shoot name " + *kymaVpc)
	} else if err != nil {
		return nil, err
	}

	//peering from kyma to remote vpc
	peeringRequestFromKyma := &pb.AddPeeringNetworkRequest{
		Network: *kymaVpc,
		Project: *kymaProject,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:                 name,
				Network:              &remoteNetwork,
				ImportCustomRoutes:   importCustomRoutes,
				ExchangeSubnetRoutes: ptr.To(true),
			},
		},
	}

	_, err = c.cnc.AddPeering(ctx, peeringRequestFromKyma)
	if err != nil {
		return nil, err
	}

	var networkPeering *pb.NetworkPeering
	net, err := c.cnc.Get(ctx, &pb.GetNetworkRequest{Network: *kymaVpc, Project: *kymaProject})
	nps := net.GetPeerings()
	for _, np := range nps {
		if *np.Network == remoteNetwork {
			networkPeering = np
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
				ExchangeSubnetRoutes: ptr.To(true),
			},
		},
	}

	_, err = c.cnc.AddPeering(ctx, peeringRequestFromRemote)
	if err != nil {
		return networkPeering, err
	}

	return networkPeering, nil
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func (c *networkClient) CheckRemoteNetworkTags(context context.Context, remoteNetwork *pb.Network, desiredTag string) (bool, error) {
	//Unfortunately get networks doesn't have the tags, so we need to use the resource manager tag bindings client
	tbc, err := resourcemanager.NewTagBindingsClient(context, option.WithCredentialsFile(abstractions.NewOSEnvironment().Get("GCP_SA_JSON_KEY_PATH")))
	if err != nil {
		return false, err
	}
	//ListEffectiveTags requires the networkId instead of name therefore we need to convert the selfLinkId to the format that the tag bindings client expects
	//more info here: https://cloud.google.com/iam/docs/full-resource-names
	tagIterator := tbc.ListEffectiveTags(context, &resourcemanagerpb.ListEffectiveTagsRequest{Parent: strings.Replace(*remoteNetwork.SelfLinkWithId, "https://www.googleapis.com/compute/v1", "//compute.googleapis.com", 1)})
	defer tbc.Close()
	for {
		tag, err := tagIterator.Next()
		if err != nil {
			return false, err
		}
		//since we are not sure where the user is going to put the tag under, let's check if the tag key contains the desired tag
		//i.e.: project/kyma-shoot-1234
		if strings.Contains(tag.NamespacedTagKey, desiredTag) {
			return true, nil
		}
	}
}
