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
creationTimestamp: '2025-12-01T02:34:04.594-08:00'
description: my router
encryptedInterconnectRouter: false
id: '1036831234123412345'
kind: compute#router
name: my-cloud-router
nats:
- autoNetworkTier: PREMIUM
  enableDynamicPortAllocation: false
  enableEndpointIndependentMapping: false
  endpointTypes:
  - ENDPOINT_TYPE_VM
  icmpIdleTimeoutSec: 30
  logConfig:
    enable: true
    filter: ERRORS_ONLY
  maxPortsPerVm: 65536
  minPortsPerVm: 2048
  name: my-cloud-nat
  natIpAllocateOption: AUTO_ONLY
  sourceSubnetworkIpRangesToNat: LIST_OF_SUBNETWORKS
  subnetworks:
  - name: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-east1/subnetworks/my-subnet
    sourceIpRangesToNat:
    - ALL_IP_RANGES
  tcpEstablishedIdleTimeoutSec: 1200
  tcpTimeWaitTimeoutSec: 120
  tcpTransitoryIdleTimeoutSec: 30
  type: PUBLIC
  udpIdleTimeoutSec: 30
network: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network
region: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-east1
selfLink: https://www.googleapis.com/compute/v1/projects/my-project/regions/us-east1/routers/my-cloud-router
*/

func (s *store) getRouterNoLock(project, region, router string) (*computepb.Router, error) {
	for _, r := range s.routers.items {
		if r.Name.ProjectId() == project && r.Name.LocationRegionId() == region && r.Name.ResourceId() == router {
			return r.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("router %s not found", gcputil.NewRouterName(project, region, router).String())
}

// Interface methods =======================================================================

func (s *store) ListRouters(ctx context.Context, req *computepb.ListRoutersRequest, opts ...gax.CallOption) gcpclient.Iterator[*computepb.Router] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Router]{err: ctx.Err()}
	}

	list := s.routers
	if req.Project != "" {
		list = list.FilterByCallback(func(item FilterableListItem[*computepb.Router]) bool {
			return item.Name.ProjectId() == req.Project
		})
	}
	if req.Region != "" {
		list = list.FilterByCallback(func(item FilterableListItem[*computepb.Router]) bool {
			return item.Name.LocationRegionId() == req.Region
		})
	}
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Router]{err: err}
	}

	return list.ToIterator()
}

func (s *store) GetRouter(ctx context.Context, req *computepb.GetRouterRequest, opts ...gax.CallOption) (*computepb.Router, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	router, err := s.getRouterNoLock(req.Project, req.Region, req.Router)
	if err != nil {
		return nil, err
	}
	return util.Clone(router)
}

func (s *store) InsertRouter(ctx context.Context, req *computepb.InsertRouterRequest, opts ...gax.CallOption) (gcpclient.VoidOperation, error) {
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
	if req.RouterResource == nil {
		return nil, gcpmeta.NewBadRequestError("router resource is required")
	}

	// network validation

	if ptr.Deref(req.RouterResource.Network, "") == "" {
		return nil, gcpmeta.NewBadRequestError("router resource must have network")
	}
	networkNd, err := gcputil.ParseNameDetail(ptr.Deref(req.RouterResource.Network, ""))
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid network reference in router resource: %v", err)
	}
	_, err = s.getNetworkNoLock(networkNd.ProjectId(), networkNd.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found for router resource", networkNd.String())
	}

	// nats subnet validation
	for _, nats := range req.RouterResource.Nats {
		if nats.GetName() == "" {
			return nil, gcpmeta.NewBadRequestError("router nats name is required")
		}
		for _, natSub := range nats.Subnetworks {
			// try first as natSub.Name is valid name
			subNd, err := gcputil.ParseNameDetail(natSub.GetName())
			if err != nil {
				subNd = gcputil.NewSubnetworkName(req.Project, req.Region, natSub.GetName())
			}
			if subNd.ResourceType() != gcputil.ResourceTypeSubnetwork {
				return nil, gcpmeta.NewBadRequestError("invalid nats subnet name %q", natSub.GetName())
			}
			_, err = s.getSubnetNoLock(req.Project, req.Region, subNd.ResourceId())
			if err != nil {
				return nil, gcpmeta.NewBadRequestError("nats subnet %s does not exist", subNd.String())
			}
			natSub.Name = ptr.To(subNd.PrefixWithGoogleApisComputeV1())
		}
	}

	// insert router

	id := rand.Uint64()
	name := gcputil.NewRouterName(req.Project, req.Region, req.RouterResource.GetName())
	router, err := util.Clone(req.RouterResource)
	if err != nil {
		return nil, fmt.Errorf("%w failed to copy router: %w", common.ErrLogical, err)
	}
	router.Id = ptr.To(id)
	router.SelfLink = ptr.To(name.PrefixWithGoogleApisComputeV1())
	router.Network = ptr.To(networkNd.ResourceId())
	router.Region = ptr.To(gcputil.NewRegionName(req.Project, req.Region).String())

	s.routers.Add(router, name)

	op := s.createComputeOperationNoLock(req.Project, req.Region, "insert", ptr.Deref(router.SelfLink, ""), id)
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))

	return newComputeOperation(op), nil
}

func (s *store) DeleteRouter(ctx context.Context, req *computepb.DeleteRouterRequest, opts ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	name := gcputil.NewRouterName(req.Project, req.Region, req.Router)
	router, err := s.getRouterNoLock(name.ProjectId(), name.LocationRegionId(), name.ResourceId())
	if err != nil {
		return nil, err
	}

	// TODO: check if router is used???

	s.routers = s.routers.FilterNotByCallback(func(item FilterableListItem[*computepb.Router]) bool {
		return item.Name.Equal(name)
	})

	op := s.createComputeOperationNoLock(req.Project, req.Region, "delete", name.PrefixWithGoogleApisComputeV1(), ptr.Deref(router.Id, 0))
	op.Status = ptr.To(computepb.Operation_DONE)
	op.EndTime = ptr.To(time.Now().Format(time.RFC3339))
	op.Progress = ptr.To(int32(100))

	return newComputeOperation(op), nil
}

// Higher level methods =======================================================================

func (s *store) GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error) {
	it := s.ListRouters(ctx, &computepb.ListRoutersRequest{
		Filter:  ptr.To(fmt.Sprintf(`network eq .*%s`, vpcName)),
		Project: project,
		Region:  region,
	}).All()
	var results []*computepb.Router
	for r, err := range it {
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
