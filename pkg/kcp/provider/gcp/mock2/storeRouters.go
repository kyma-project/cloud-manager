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
	"github.com/mitchellh/copystructure"
	"k8s.io/utils/ptr"
)

func (s *store) getRouterNoLock(project, region, router string) (*computepb.Router, error) {
	for _, r := range s.routers.items {
		if r.name.ProjectId() == project && r.name.LocationRegionId() == region && r.name.ResourceId() == router {
			return r.obj, nil
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
		list = list.filterByCallback(func(item listItem[*computepb.Router]) bool {
			return item.name.ProjectId() == req.Project
		})
	}
	if req.Region != "" {
		list = list.filterByCallback(func(item listItem[*computepb.Router]) bool {
			return item.name.LocationRegionId() == req.Region
		})
	}
	var err error
	list, err = list.filterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Router]{err: err}
	}

	return list.toIterator()
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
	cpy, err := copystructure.Copy(router)
	if err != nil {
		return nil, fmt.Errorf("failed to copy router: %w", err)
	}
	return cpy.(*computepb.Router), nil
}

func (s *store) InsertRouter(ctx context.Context, req *computepb.InsertRouterRequest, opts ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
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

	id := rand.Uint64()
	name := gcputil.NewRouterName(req.Project, req.Region, req.RouterResource.GetName())
	routerIntf, err := copystructure.Copy(req.RouterResource)
	if err != nil {
		return nil, fmt.Errorf("%w failed to copy router: %w", common.ErrLogical, err)
	}
	router := routerIntf.(*computepb.Router)
	router.Id = ptr.To(id)
	router.SelfLink = ptr.To(name.PrefixWithGoogleApisComputeV1())

	s.routers.add(router, name)

	op := s.createComputeOperationNoLock(req.Project, req.Region, "insert", ptr.Deref(router.SelfLink, ""), id)
	return newVoidOperationFromComputeOperation(op), nil
}

func (s *store) DeleteRouter(ctx context.Context, req *computepb.DeleteRouterRequest, opts ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
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

	s.routers = s.routers.filterNotByCallback(func(item listItem[*computepb.Router]) bool {
		return item.name.Equal(name)
	})

	op := s.createComputeOperationNoLock(req.Project, req.Region, "delete", name.PrefixWithGoogleApisComputeV1(), ptr.Deref(router.Id, 0))
	return newVoidOperationFromComputeOperation(op), nil
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
