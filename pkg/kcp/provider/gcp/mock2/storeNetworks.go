package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

var _ gcpclient.NetworkClient = (*store)(nil)

func (s *store) getNetworkNoLock(project, network string) (*computepb.Network, error) {
	for _, n := range s.networks.items {
		if n.Name.ProjectId() == project && n.Name.ResourceId() == network {
			return n.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("network %s not found", gcputil.NewGlobalNetworkName(project, network).String())
}

// NetworkClient interface methods ============================================================

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

	net, err := s.getNetworkNoLock(req.Project, req.Network)
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
