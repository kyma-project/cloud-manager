package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/elliotchance/pie/v2"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

/*
autoCreateSubnetworks: false
creationTimestamp: '2024-10-23T06:21:43.842-07:00'
description: ''
id: '6812341911334112200'
kind: compute#network
mtu: 1460
name: my-network
networkFirewallPolicyEnforcementOrder: AFTER_CLASSIC_FIREWALL
peerings:
- autoCreateRoutes: true
  connectionStatus:
    trafficConfiguration:
      exportCustomRoutesToPeer: false
      exportSubnetRoutesWithPublicIpToPeer: false
      importCustomRoutesFromPeer: false
      importSubnetRoutesWithPublicIpFromPeer: false
      stackType: IPV4_ONLY
    updateStrategy: INDEPENDENT
  exchangeSubnetRoutes: true
  exportCustomRoutes: false
  exportSubnetRoutesWithPublicIp: true
  importCustomRoutes: false
  importSubnetRoutesWithPublicIp: false
  name: my-peering
  network: https://www.googleapis.com/compute/v1/projects/my-remote-project/global/networks/my-remote-network
  stackType: IPV4_ONLY
  state: ACTIVE
  stateDetails: '[2026-03-02T14:50:39.800-08:00]: Connected.'
  updateStrategy: INDEPENDENT
routingConfig:
  bgpBestPathSelectionMode: LEGACY
  routingMode: REGIONAL
selfLink: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network
selfLinkWithId: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/6812341911334112200
subnetworks:
- https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/subnetworks/my-subnet
x_gcloud_bgp_routing_mode: REGIONAL
x_gcloud_subnet_mode: CUSTOM
*/

type VpcNetworkConfig interface {
	GetNetworkNoLock(project, network string) (*computepb.Network, error)
	GetNetworkAddressSpaceNoLock(netName string) *AddressSpace
}

var _ gcpclient.VpcNetworkClient = (*store)(nil)

func (s *store) GetNetworkNoLock(project, network string) (*computepb.Network, error) {
	for _, n := range s.networks.items {
		if n.Name.ProjectId() == project && n.Name.ResourceId() == network {
			return n.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("network %s not found", gcputil.NewGlobalNetworkName(project, network).String())
}

func (s *store) GetNetworkAddressSpaceNoLock(netName string) *AddressSpace {
	nd, err := gcputil.ParseNameDetail(netName)
	if err != nil {
		return nil
	}
	return s.addressSpaces[nd.String()]
}

// VpcNetworkClient interface methods ============================================================

func (s *store) GetNetwork(ctx context.Context, req *computepb.GetNetworkRequest, _ ...gax.CallOption) (*computepb.Network, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	net, err := s.GetNetworkNoLock(req.Project, req.Network)
	if err != nil {
		return nil, err
	}
	return util.Clone(net)
}

func (s *store) InsertNetwork(ctx context.Context, req *computepb.InsertNetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Project == "" {
		return nil, gcpmeta.NewBadRequestError("project is required")
	}
	if req.NetworkResource == nil {
		return nil, gcpmeta.NewBadRequestError("network resource is required")
	}
	if req.NetworkResource.Name == nil {
		return nil, gcpmeta.NewBadRequestError("network name is required")
	}
	if _, err := s.GetNetworkNoLock(req.Project, req.NetworkResource.GetName()); err == nil {
		return nil, gcpmeta.NewBadRequestError("network %s already exists", gcputil.NewGlobalNetworkName(req.Project, req.NetworkResource.GetName()).String())
	}
	if req.NetworkResource.Subnetworks != nil {
		return nil, gcpmeta.NewBadRequestError("subnetworks field is not supported for network creation")
	}

	net, err := util.JsonClone(req.NetworkResource)
	if err != nil {
		return nil, fmt.Errorf("failed to clone network resource: %w", err)
	}

	id := rand.Uint64()
	name := gcputil.NewGlobalNetworkName(req.Project, req.NetworkResource.GetName())
	net.Id = ptr.To(id)
	net.SelfLink = ptr.To(name.PrefixWithGoogleApisComputeV1())
	net.SelfLinkWithId = ptr.To(gcputil.NewGlobalNetworkName(req.Project, fmt.Sprintf("%d", id)).PrefixWithGoogleApisComputeV1())
	net.Kind = ptr.To("compute#network")

	s.addressSpaces[name.String()] = newAddressSpace()
	s.networks.Add(net, name)

	compOp := s.createComputeOperationNoLock(req.Project, "", "insert", net.GetSelfLink(), id)
	compOp.Status = ptr.To(computepb.Operation_DONE)
	compOp.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	compOp.Progress = ptr.To(int32(100))

	return newComputeOperation(compOp), nil
}

func (s *store) ListNetworks(ctx context.Context, req *computepb.ListNetworksRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Network] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Network]{err: ctx.Err()}
	}

	list := s.networks
	if req.Project != "" {
		list = s.networks.FilterByCallback(func(l FilterableListItem[*computepb.Network]) bool {
			return l.Name.ProjectId() == req.Project
		})
	}
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Network]{err: fmt.Errorf("failed to filter networks by expression: %w", err)}
	}

	return list.ToIterator()
}

func (s *store) DeleteNetwork(ctx context.Context, req *computepb.DeleteNetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	net, err := s.GetNetworkNoLock(req.Project, req.Network)
	if err != nil {
		return nil, err
	}

	nd := gcputil.NewGlobalNetworkName(req.Project, req.Network)
	ndTxt := nd.String()

	// check if network contains some resources

	for _, item := range s.subnets.items {
		if strings.Contains(item.Obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has subnet %s", nd.String(), gcputil.MustParseNameDetail(item.Obj.GetSelfLink()).String())
		}
	}
	for _, item := range s.routers.items {
		if strings.Contains(item.Obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has router %s", nd.String(), gcputil.MustParseNameDetail(item.Obj.GetSelfLink()).String())
		}
	}
	for _, item := range s.addresses.items {
		if strings.Contains(item.Obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has address %s", nd.String(), gcputil.MustParseNameDetail(item.Obj.GetSelfLink()).String())
		}
	}

	// remove the network

	s.networks = s.networks.FilterNotByCallback(func(item FilterableListItem[*computepb.Network]) bool {
		return item.Name.Equal(nd)
	})
	delete(s.addressSpaces, ndTxt)

	compOp := s.createComputeOperationNoLock(req.Project, "", "delete", nd.PrefixWithGoogleApisComputeV1(), ptr.Deref(net.Id, 0))
	compOp.Status = ptr.To(computepb.Operation_DONE)
	compOp.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	compOp.Progress = ptr.To(int32(100))

	return newComputeOperation(compOp), nil
}

func (s *store) AddPeering(ctx context.Context, req *computepb.AddPeeringNetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netLocalNd := gcputil.NewGlobalNetworkName(req.Project, req.Network)
	netLocal, err := s.GetNetworkNoLock(req.Project, req.Network)
	if err != nil {
		return nil, err
	}
	if req.NetworksAddPeeringRequestResource == nil {
		return nil, gcpmeta.NewBadRequestError("networksAddPeeringRequestResource is required")
	}
	if req.NetworksAddPeeringRequestResource.NetworkPeering == nil {
		return nil, gcpmeta.NewBadRequestError("networkPeering is required")
	}
	if req.NetworksAddPeeringRequestResource.NetworkPeering.GetNetwork() == "" {
		return nil, gcpmeta.NewBadRequestError("network is required")
	}

	netRemoteNd, err := gcputil.ParseNameDetail(req.NetworksAddPeeringRequestResource.NetworkPeering.GetNetwork())
	if err != nil {
		netRemoteNd = gcputil.NewGlobalNetworkName(req.Project, req.NetworksAddPeeringRequestResource.NetworkPeering.GetNetwork())
	}

	storeRemote := s.server.GetSubscription(netRemoteNd.ProjectId())
	if storeRemote == nil {
		return nil, gcpmeta.NewInternalServerError("remote subscription %s not found", netRemoteNd.ProjectId())
	}

	netRemote, err := storeRemote.GetNetworkNoLock(storeRemote.ProjectId(), netRemoteNd.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("remote network %s not found", netRemoteNd.String())
	}

	// address space overlap check, symbolic
	// this is a simplified check, since other peerings are not investigated, only local subnets and address ranges
	// if remote network does not exist, then its address space is also not found

	asLocal := s.addressSpaces[netLocalNd.String()]
	if asLocal == nil {
		return nil, gcpmeta.NewInternalServerError("%v: unexpected can not find local network %s address space", common.ErrLogical, netLocalNd.String())
	}
	asRemote := storeRemote.GetNetworkAddressSpaceNoLock(netRemoteNd.ProjectId())
	if asLocal.Overlaps(asRemote) {
		return nil, gcpmeta.NewBadRequestError("address space overlaps %s - %s", asLocal.String(), asRemote.String())
	}

	// check if already exists:
	// * local peering to remote network
	// * local peering with same name

	for _, localPeering := range netLocal.Peerings {
		if netRemoteNd.EqualString(localPeering.GetNetwork()) {
			return nil, gcpmeta.NewBadRequestError("local network already has peering to remote network %s", netRemoteNd.String())
		}
		if localPeering.GetName() == req.NetworksAddPeeringRequestResource.NetworkPeering.GetName() {
			return nil, gcpmeta.NewBadRequestError("local network already has peering with name %s", req.NetworksAddPeeringRequestResource.NetworkPeering.GetName())
		}
	}

	peeringLocal, err := util.Clone(req.NetworksAddPeeringRequestResource.NetworkPeering)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to clone peering: %v", common.ErrLogical, err)
	}
	peeringLocal.Network = ptr.To(netRemoteNd.PrefixWithGoogleApisComputeV1())
	peeringLocal.State = ptr.To(computepb.NetworkPeering_INACTIVE.String())
	netLocal.Peerings = append(netLocal.Peerings, peeringLocal)

	// look for remote peering to establish connection
	var peeringRemote *computepb.NetworkPeering
	for _, peering := range netRemote.Peerings {
		if netLocalNd.EqualString(peering.GetNetwork()) {
			peeringRemote = peering
			break
		}
	}
	// establish the connection, since peerings exist on both sides
	if peeringRemote != nil {
		peeringRemote.State = ptr.To(computepb.NetworkPeering_ACTIVE.String())
		peeringLocal.State = ptr.To(computepb.NetworkPeering_ACTIVE.String())
	}

	op := s.createComputeOperationNoLock(s.projectId, "", "addPeering", netLocalNd.PrefixWithGoogleApisComputeV1(), netLocal.GetId())
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))

	return newComputeOperation(op), nil
}

func (s *store) RemovePeering(ctx context.Context, req *computepb.RemovePeeringNetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netLocalNd := gcputil.NewGlobalNetworkName(req.Project, req.Network)
	netLocal, err := s.GetNetworkNoLock(netLocalNd.ProjectId(), netLocalNd.ResourceId())
	if err != nil {
		return nil, err
	}
	if req.NetworksRemovePeeringRequestResource == nil {
		return nil, gcpmeta.NewBadRequestError("networksRemovePeeringRequestResource is required")
	}
	if req.NetworksRemovePeeringRequestResource.GetName() == "" {
		return nil, gcpmeta.NewBadRequestError("peering name is required")
	}

	// remove local peering

	found := false
	var netRemoteSelfLink string
	netLocal.Peerings = pie.FilterNot(netLocal.Peerings, func(peering *computepb.NetworkPeering) bool {
		if peering.GetName() == req.NetworksRemovePeeringRequestResource.GetName() {
			found = true
			netRemoteSelfLink = peering.GetNetwork()
			return true
		}
		return false
	})
	if !found {
		return nil, gcpmeta.NewBadRequestError("peering %s not found in network %s", req.NetworksRemovePeeringRequestResource.GetName(), netLocalNd.String())
	}

	// look for remote peering and update its status

	netRemoteNd, err := gcputil.ParseNameDetail(netRemoteSelfLink)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to parse remote network self link: %q", common.ErrLogical, netRemoteSelfLink)
	}

	storeRemote := s.server.GetSubscription(netRemoteNd.ProjectId())
	if storeRemote == nil {
		return nil, gcpmeta.NewInternalServerError("%v: remote store %s not found", common.ErrLogical, netRemoteNd.ProjectId())
	}

	netRemote, err := storeRemote.GetNetworkNoLock(netRemoteNd.ProjectId(), netRemoteNd.ResourceId())
	if err == nil {
		for _, peering := range netRemote.Peerings {
			if netLocalNd.EqualString(peering.GetNetwork()) {
				peering.State = ptr.To(computepb.NetworkPeering_INACTIVE.String())
			}
		}
	}

	op := s.createComputeOperationNoLock(s.projectId, "", "removePeering", netLocal.GetSelfLink(), netLocal.GetId())
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))

	return newComputeOperation(op), nil
}
