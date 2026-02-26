package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/mitchellh/copystructure"
	"k8s.io/utils/ptr"
)

var _ gcpclient.NetworkClient = (*store)(nil)

func (s *store) getNetworkNoLock(project, network string) (*computepb.Network, error) {
	for _, n := range s.networks.items {
		if n.name.ProjectId() == project && n.name.ResourceId() == network {
			return n.obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("network %s not found", gcputil.NewGlobalNetworkName(project, network).String())
}

func (s *store) GetNetwork(ctx context.Context, req *computepb.GetNetworkRequest, _ ...gax.CallOption) (*computepb.Network, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	net, err := s.getNetworkNoLock(req.Project, req.Network)
	if err != nil {
		return nil, err
	}
	cpy, err := copystructure.Copy(net)
	if err != nil {
		return nil, fmt.Errorf("failed to copy network: %w", err)
	}
	return cpy.(*computepb.Network), nil
}

func (s *store) InsertNetwork(ctx context.Context, req *computepb.InsertNetworkRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
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
	if _, err := s.getNetworkNoLock(req.Project, req.NetworkResource.GetName()); err == nil {
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

	s.addressSpaces[net.GetSelfLink()] = newAddressSpace()
	s.networks.add(net, name)

	compOp := s.createComputeOperationNoLock(req.Project, "", "insert", net.GetSelfLink(), id)

	return newVoidOperationFromComputeOperation(compOp), nil
}

func (s *store) ListNetworks(ctx context.Context, req *computepb.ListNetworksRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Network] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Network]{err: ctx.Err()}
	}

	list := s.networks
	if req.Project != "" {
		list = s.networks.filterByCallback(func(l listItem[*computepb.Network]) bool {
			return l.name.ProjectId() == req.Project
		})
	}
	var err error
	list, err = list.filterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Network]{err: fmt.Errorf("failed to filter networks by expression: %w", err)}
	}

	return list.toIterator()
}

func (s *store) DeleteNetwork(ctx context.Context, req *computepb.DeleteNetworkRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	net, err := s.getNetworkNoLock(req.Project, req.Network)
	if err != nil {
		return nil, err
	}

	nd := gcputil.NewGlobalNetworkName(req.Project, req.Network)
	ndTxt := nd.String()

	// check if network contains some resources

	for _, subnet := range s.subnets.items {
		if strings.Contains(subnet.obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has subnet %s", nd.String(), subnet.obj.GetSelfLink())
		}
	}
	for _, router := range s.routers.items {
		if strings.Contains(router.obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has router %s", nd.String(), router.obj.GetSelfLink())
		}
	}
	for _, router := range s.addresses.items {
		if strings.Contains(router.obj.GetNetwork(), ndTxt) {
			return nil, gcpmeta.NewBadRequestError("network %s cannot be deleted because it has address %s", nd.String(), router.obj.GetSelfLink())
		}
	}

	// remove the network

	s.networks = s.networks.filterNotByCallback(func(item listItem[*computepb.Network]) bool {
		return item.name.Equal(nd)
	})
	delete(s.addressSpaces, ndTxt)

	compOp := s.createComputeOperationNoLock(req.Project, "", "delete", nd.PrefixWithGoogleApisComputeV1(), ptr.Deref(net.Id, 0))
	return newVoidOperationFromComputeOperation(compOp), nil
}
