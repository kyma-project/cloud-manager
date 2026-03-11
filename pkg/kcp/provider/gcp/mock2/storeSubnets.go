package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

/*
allowSubnetCidrRoutesOverlap: false
creationTimestamp: '2019-07-16T01:28:03.808-07:00'
enableFlowLogs: true
fingerprint: Z1234gLxMMQ=
gatewayAddress: 10.128.0.1
id: '812349123460231412'
ipCidrRange: 10.128.0.0/20
kind: compute#subnetwork
logConfig:
  aggregationInterval: INTERVAL_5_SEC
  enable: true
  flowSampling: 0.5
  metadata: INCLUDE_ALL_METADATA
name: my-subnet
network: https://www.googleapis.com/compute/v1/projects/my-network/global/networks/my-subnet
privateIpGoogleAccess: true
privateIpv6GoogleAccess: DISABLE_GOOGLE_ACCESS
purpose: PRIVATE
region: https://www.googleapis.com/compute/v1/projects/my-network/regions/us-central1
selfLink: https://www.googleapis.com/compute/v1/projects/my-network/regions/us-central1/subnetworks/my-subnet
stackType: IPV4_ONLY
 */

func (s *store) getSubnetNoLock(project, region, subnet string) (*computepb.Subnetwork, error) {
	nd := gcputil.NewSubnetworkName(project, region, subnet)
	result, found := s.subnets.FindByName(nd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("subnet %s not found", gcputil.NewSubnetworkName(project, region, subnet).String())
	}
	return result, nil
}

func (s *store) InsertSubnet(ctx context.Context, req *computepb.InsertSubnetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
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
	if req.SubnetworkResource == nil {
		return nil, gcpmeta.NewBadRequestError("subnetwork resource is required")
	}

	// subnet name uniqueness
	_, err := s.getSubnetNoLock(req.Project, req.Region, req.SubnetworkResource.GetName())
	if err == nil {
		return nil, gcpmeta.NewBadRequestError("subnet %s already exists", gcputil.NewSubnetworkName(req.Project, req.Region, req.SubnetworkResource.GetName()).String())
	}

	// network
	if req.SubnetworkResource.GetNetwork() == "" {
		return nil, gcpmeta.NewBadRequestError("network is required for subnetwork creation")
	}
	networkName, err := gcputil.ParseNameDetail(req.SubnetworkResource.GetNetwork())
	if err != nil {
		networkName = gcputil.NewGlobalNetworkName(req.Project, req.SubnetworkResource.GetNetwork())
	}
	if networkName.ResourceType() != gcputil.ResourceTypeGlobalNetwork {
		return nil, gcpmeta.NewBadRequestError("invalid subnet network name type")
	}
	net, err := s.getNetworkNoLock(networkName.ProjectId(), networkName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found", networkName.String())
	}

	// cidr & network address space checks

	addressSpace, ok := s.addressSpaces[networkName.String()]
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
	subnet.Region = ptr.To(req.Region)
	subnet.State = ptr.To(computepb.Subnetwork_READY.String())
	subnet.Network = ptr.To(networkName.PrefixWithGoogleApisComputeV1())

	net.Subnetworks = append(net.Subnetworks, subnet.GetSelfLink())

	if err := addressSpace.Reserve(subnetRanges...); err != nil {
		return nil, fmt.Errorf("%w failed to add already verified subnet ranges: %w", common.ErrLogical, err)
	}
	s.subnets.Add(subnet, name)

	op := s.createComputeOperationNoLock(req.Project, req.Region, "insert", subnet.GetSelfLink(), id)
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))

	return newComputeOperation(op), nil
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

func (s *store) ListSubnets(ctx context.Context, req *computepb.ListSubnetworksRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Subnetwork] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Subnetwork]{
			err: ctx.Err(),
		}
	}

	var err error

	list := s.subnets
	if req.Project != "" {
		list = list.FilterNotByCallback(func(item FilterableListItem[*computepb.Subnetwork]) bool {
			return item.Name.ProjectId() == req.Project
		})
	}
	if req.Region != "" {
		list = list.FilterNotByCallback(func(item FilterableListItem[*computepb.Subnetwork]) bool {
			return item.Name.LocationRegionId() == req.Region
		})
	}
	if req.Filter != nil {
		list, err = list.FilterByExpression(req.Filter)
		if err != nil {
			return &iteratorMocked[*computepb.Subnetwork]{
				err: gcpmeta.NewBadRequestError("failed to filter by expression: %v", err),
			}
		}
	}

	return list.ToIterator()
}

func (s *store) DeleteSubnet(ctx context.Context, req *computepb.DeleteSubnetworkRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	subNd := gcputil.NewSubnetworkName(req.Project, req.Region, req.Subnetwork)

	sub, err := s.getSubnetNoLock(req.Project, req.Region, req.Subnetwork)
	if err != nil {
		return nil, err
	}

	for _, item := range s.routers.items {
		for _, nats := range item.Obj.Nats {
			for _, natSub := range nats.Subnetworks {
				if subNd.EqualString(natSub.GetName()) {
					return nil, gcpmeta.NewBadRequestError("subnet %s is attached to router %s", subNd.String(), item.Name.String())
				}
			}
		}
	}
	for _, item := range s.serviceConnectionPolicies.items {
		if item.Obj.PscConfig != nil {
			for _, subTxt := range item.Obj.PscConfig.Subnetworks {
				scpSubnetName, err := gcputil.ParseNameDetail(subTxt)
				if err != nil {
					return nil, fmt.Errorf("%w: SCP %s subnet has invalid name %s: %w", common.ErrLogical, item.Name.String(), subTxt, err)
				}
				if scpSubnetName.Equal(subNd) {
					return nil, gcpmeta.NewBadRequestError("subnet %s is attached to SCP %s", scpSubnetName.String(), item.Name.String())
				}
			}
		}
	}
	// add additional checks for other existing resources using this subnet

	// remove the subnet

	s.subnets = s.subnets.FilterNotByCallback(func(item FilterableListItem[*computepb.Subnetwork]) bool {
		return item.Name.Equal(subNd)
	})

	compOp := s.createComputeOperationNoLock(req.Project, req.Region, "delete", subNd.PrefixWithGoogleApisComputeV1(), sub.GetId())
	compOp.Status = ptr.To(computepb.Operation_DONE)
	compOp.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	compOp.Progress = ptr.To(int32(100))

	return newComputeOperation(compOp), nil
}
