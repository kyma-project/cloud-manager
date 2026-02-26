package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func (s *store) getSubnetNoLock(project, region, subnet string) (*computepb.Subnetwork, error) {
	nd := gcputil.NewSubnetworkName(project, region, subnet)
	result, found := s.subnets.findByName(nd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("subnet %s not found", gcputil.NewSubnetworkName(project, region, subnet).String())
	}
	return result, nil
}

func (s *store) InsertSubnet(ctx context.Context, req *computepb.InsertSubnetworkRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.SubnetworkResource == nil {
		return nil, gcpmeta.NewBadRequestError("subnetwork resource is required")
	}

	// subnet name uniqueness
	_, err := s.getSubnetNoLock(req.Project, req.Region, req.SubnetworkResource.GetName())
	if err == nil {
		return nil, gcpmeta.NewBadRequestError("subnet %s already exists", gcputil.NewSubnetworkName(req.Project, req.Region, req.SubnetworkResource.GetName()).String())
	}

	// network
	if ptr.Deref(req.SubnetworkResource.Network, "") == "" {
		return nil, gcpmeta.NewBadRequestError("network is required for subnetwork creation")
	}
	networkName, err := gcputil.ParseNameDetail(ptr.Deref(req.SubnetworkResource.Network, ""))
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid network name: %v", err)
	}
	net, err := s.getNetworkNoLock(networkName.ProjectId(), networkName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found", networkName.String())
	}

	// cidr & network address space checks

	addressSpace, ok := s.addressSpaces[net.GetSelfLink()]
	if !ok {
		return nil, fmt.Errorf("%w address space for network %s not found", common.ErrLogical, net.GetSelfLink())
	}
	var subnetRanges []string
	// IMPORTANT! Check is performed on a clone of the address space, so that if there is an error,
	// the original address space is not modified and can be used for correct error reporting in subsequent calls
	if subnetRanges, err = addressSpace.Clone().AddSubnet(req.SubnetworkResource); err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid subnet ip cidr range(s): %v", err)
	}

	// create the subnet

	subnet, err := util.JsonClone(req.SubnetworkResource)
	if err != nil {
		return nil, fmt.Errorf("failed to clone subnetwork resource: %w", err)
	}

	id := rand.Uint64()
	name := gcputil.NewSubnetworkName(req.Project, req.Region, subnet.GetName())
	subnet.Id = ptr.To(id)
	subnet.Kind = ptr.To("compute#subnetwork")
	subnet.SelfLink = ptr.To(name.PrefixWithGoogleApisComputeV1())

	net.Subnetworks = append(net.Subnetworks, ptr.Deref(subnet.SelfLink, ""))

	if err := addressSpace.Reserve(subnetRanges...); err != nil {
		return nil, fmt.Errorf("%w failed to add already verified subnet ranges: %w", common.ErrLogical, err)
	}
	s.subnets.add(subnet, name)

	op := s.createComputeOperationNoLock(req.Project, req.Region, "insert", ptr.Deref(subnet.SelfLink, ""), id)
	return newVoidOperationFromComputeOperation(op), nil
}

func (s *store) GetSubnet(ctx context.Context, req *computepb.GetSubnetworkRequest, _ ...gax.CallOption) (*computepb.Subnetwork, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Project == "" {
		return nil, gcpmeta.NewBadRequestError("project is required")
	}
	if req.Region == "" {
		return nil, gcpmeta.NewBadRequestError("region is required")
	}
	if req.Subnetwork == "" {
		return nil, gcpmeta.NewBadRequestError("subnetwork is required")
	}

	subnet, err := s.getSubnetNoLock(req.Project, req.Region, req.Subnetwork)
	if err != nil {
		return nil, err
	}

	cpy, err := util.JsonClone(subnet)
	if err != nil {
		return nil, fmt.Errorf("failed to clone subnet: %w", err)
	}

	return cpy, nil
}

func (s *store) DeleteSubnet(ctx context.Context, req *computepb.DeleteSubnetworkRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	sub, err := s.getSubnetNoLock(req.Project, req.Region, req.Subnetwork)
	if err != nil {
		return nil, err
	}

	// TODO: check for existing resources using this subnet

	nd := gcputil.NewSubnetworkName(req.Project, req.Region, req.Subnetwork)
	//ndTxt := nd.String()

	// not sure how to check if there are redis clusters using this subnet, they need a PSC connection and a policy linked to the subnet
	// TODO: implement the check if subnet is used by PSC connection policy

	// remove the subnet

	s.subnets = s.subnets.filterNotByCallback(func(item listItem[*computepb.Subnetwork]) bool {
		return item.name.Equal(nd)
	})

	compOp := s.createComputeOperationNoLock(req.Project, req.Region, "delete", nd.PrefixWithGoogleApisComputeV1(), ptr.Deref(sub.Id, 0))
	return newVoidOperationFromComputeOperation(compOp), nil
}
