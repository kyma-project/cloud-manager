/*
required GCP permissions
=========================
  - Both Sides - The service account used to create the VPC peering connection needs the following permissions:
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Removes the VPC peering connection
  compute.networks.removePeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/removePeering
  ** Gets the network (VPCs) in order to retrieve all peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get

  - Remote Side - The service account used to create the VPC peering connection needs the additional permissions:
  ** Fetches the remote network tags
  compute.networks.ListEffectiveTags => https://cloud.google.com/resource-manager/reference/rest/v3/tagKeys/get
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/api/iterator"

	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"k8s.io/utils/ptr"
)

func NewClientProvider(gcpClients *client.GcpClients) client.GcpClientProvider[VpcPeeringClient] {
	return func() VpcPeeringClient { return NewVpcPeeringClient(gcpClients) }
}

func NewVpcPeeringClient(gcpClients *client.GcpClients) VpcPeeringClient {
	if gcpClients.VpcPeeringClients == nil {
		return &gcpVpcPeeringClient{}
	}
	return &gcpVpcPeeringClient{networksClient: gcpClients.VpcPeeringClients.ComputeNetworks, resourceManagerClient: gcpClients.VpcPeeringClients.ResourceManagerTagBindings}
}

type gcpVpcPeeringClient struct {
	networksClient        *compute.NetworksClient
	resourceManagerClient *resourcemanager.TagBindingsClient
}

type VpcPeeringClient interface {
	DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) error
	GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error)
	CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) error
	CreateKymaVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) error
	GetRemoteNetworkTags(context context.Context, remoteVpc string, remoteProject string) ([]string, error)
}

func CreateVpcPeeringRequest(ctx context.Context, remotePeeringName string, sourceVpc string, sourceProject string, importCustomRoutes bool, exportCustomRoutes bool, destinationProject string, destinationVpc string, networksClient *compute.NetworksClient) error {

	destinationNetworkUrl := getFullNetworkUrl(destinationProject, destinationVpc)

	vpcPeeringRequest := &pb.AddPeeringNetworkRequest{
		Network: sourceVpc,
		Project: sourceProject,
		NetworksAddPeeringRequestResource: &pb.NetworksAddPeeringRequest{
			NetworkPeering: &pb.NetworkPeering{
				Name:                 &remotePeeringName,
				Network:              &destinationNetworkUrl,
				ExportCustomRoutes:   &exportCustomRoutes,
				ExchangeSubnetRoutes: ptr.To(true),
				ImportCustomRoutes:   &importCustomRoutes,
			},
		},
	}

	_, err := networksClient.AddPeering(ctx, vpcPeeringRequest)
	if err != nil {
		return err
	}
	return nil

}

func (c *gcpVpcPeeringClient) CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) error {
	//peering from remote vpc to kyma
	//by default exportCustomRoutes is false, but if the remote vpc wants kyma to import custom routes, the peering needs to export them :)
	exportCustomRoutes := false
	importCustomRoutes := false
	if customRoutes {
		exportCustomRoutes = true
	}
	return CreateVpcPeeringRequest(ctx, remotePeeringName, remoteVpc, remoteProject, importCustomRoutes, exportCustomRoutes, kymaProject, kymaVpc, c.networksClient)
}

func (c *gcpVpcPeeringClient) CreateKymaVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) error {
	//peering from kyma to remote vpc
	//Kyma will not export custom routes to the remote vpc, but if the remote vpc is exporting them, we need to import them
	exportCustomRoutes := false
	importCustomRoutes := customRoutes
	return CreateVpcPeeringRequest(ctx, remotePeeringName, kymaVpc, kymaProject, importCustomRoutes, exportCustomRoutes, remoteProject, remoteVpc, c.networksClient)
}

func (c *gcpVpcPeeringClient) DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) error {
	_, err := c.networksClient.RemovePeering(ctx, &pb.RemovePeeringNetworkRequest{
		Network:                              kymaVpc,
		Project:                              kymaProject,
		NetworksRemovePeeringRequestResource: &pb.NetworksRemovePeeringRequest{Name: &remotePeeringName},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *gcpVpcPeeringClient) GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error) {

	network, err := c.networksClient.Get(ctx, &pb.GetNetworkRequest{Network: vpc, Project: project})
	if err != nil {
		return nil, err
	}
	peerings := pie.Filter(network.GetPeerings(), func(peering *pb.NetworkPeering) bool { return peering.GetName() == remotePeeringName })

	if len(peerings) == 0 {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Vpc Peering not found")
		return nil, nil
	}
	return peerings[0], nil
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func (c *gcpVpcPeeringClient) GetRemoteNetworkTags(ctx context.Context, remoteVpc string, remoteProject string) ([]string, error) {
	var tagsArray []string

	//NetworkPeering will only be created if the remote vpc has a tag with the kyma shoot name
	remoteNetwork, err := c.networksClient.Get(ctx, &pb.GetNetworkRequest{Network: remoteVpc, Project: remoteProject})
	if err != nil {
		return nil, err
	}

	tagIterator := c.resourceManagerClient.ListEffectiveTags(ctx, &resourcemanagerpb.ListEffectiveTagsRequest{Parent: strings.Replace(ptr.Deref(remoteNetwork.SelfLinkWithId, ""), "https://www.googleapis.com/compute/v1", "//compute.googleapis.com", 1)})
	for {
		tag, err := tagIterator.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		tagsArray = append(tagsArray, tag.NamespacedTagKey)
	}
	return tagsArray, nil
}
