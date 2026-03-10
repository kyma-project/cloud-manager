package mock2

import (
	"context"
	"fmt"

	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

/*
createTime: '2026-01-29T09:32:47.466418964Z'
description: lorem ipsum
etag: zWkgLkXas3-XWiArzWeQwUNJmEd32KIdMDDYSFreHs0
infrastructure: PSC
name: projects/my-project/locations/us-east1/serviceConnectionPolicies/my-policy
network: projects/my-project/global/networks/my-network
pscConfig:
  subnetworks:
  - projects/my-project/regions/us-east1/subnetworks/my-subnet
serviceClass: gcp-memorystore-redis
updateTime: '2026-01-29T09:32:56.602453767Z'
*/

func (s *store) getServiceConnectionPolicyNoLock(scpFullName string) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	for _, item := range s.serviceConnectionPolicies.items {
		if item.Name.EqualString(scpFullName) {
			return item.Obj, nil
		}
	}
	return nil, gcpmeta.NewNotFoundError("serviceConnectionPolicy %s not found", scpFullName)
}

// NetworkConnectivityClient public methods

func (s *store) CreateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.CreateServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	s.m.Lock()
	s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Parent == "" {
		return nil, gcpmeta.NewBadRequestError("parent is required")
	}
	parentName, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("parent is invalid: %v", err)
	}
	if req.ServiceConnectionPolicyId == "" {
		return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicyId is required")
	}
	if req.ServiceConnectionPolicy == nil {
		return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicy is required")
	}

	scpName := gcputil.NewServiceConnectionPolicyName(parentName.ProjectId(), parentName.LocationRegionId(), req.ServiceConnectionPolicyId)

	if req.ServiceConnectionPolicy.ServiceClass == "" {
		return nil, gcpmeta.NewBadRequestError("serviceClass is required")
	}

	// validate network exists

	netName, err := gcputil.ParseNameDetail(req.ServiceConnectionPolicy.Network)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network name is invalid: %v", err)
	}
	if netName.ResourceType() != gcputil.ResourceTypeGlobalNetwork {
		return nil, gcpmeta.NewBadRequestError("network name type expected global network, but got %s", netName.ResourceType())
	}

	_, err = s.getNetworkNoLock(netName.ProjectId(), netName.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s does not exist: %v", netName.String(), err)
	}

	// validate there's no other SCP with the same network, region and serviceClass

	for _, item := range s.serviceConnectionPolicies.items {
		if item.Name.ProjectId() != parentName.ProjectId() || item.Name.LocationRegionId() != parentName.LocationRegionId() {
			continue
		}
		if !netName.EqualString(item.Obj.Network) {
			continue
		}
		if item.Obj.ServiceClass == req.ServiceConnectionPolicy.ServiceClass {
			return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicy with network %s, region %s and serviceClass %s already exist", netName.String(), parentName.LocationRegionId(), req.ServiceConnectionPolicy.ServiceClass)
		}
	}

	// validate subnets exist

	if len(req.ServiceConnectionPolicy.PscConfig.Subnetworks) == 0 {
		return nil, gcpmeta.NewBadRequestError("subnetwork is required")
	}

	for _, subNetNameTxt := range req.ServiceConnectionPolicy.PscConfig.Subnetworks {
		subnetName, err := gcputil.ParseNameDetail(subNetNameTxt)
		if err != nil {
			return nil, gcpmeta.NewBadRequestError("subnetwork name is invalid: %v", err)
		}
		_, err = s.getSubnetNoLock(subnetName.ProjectId(), subnetName.LocationRegionId(), subnetName.ResourceId())
		if err != nil {
			return nil, gcpmeta.NewBadRequestError("subnetwork %s does not exist: %v", subnetName.String(), err)
		}
	}

	scp, err := util.Clone(req.ServiceConnectionPolicy)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v failed to clone service connection policy: %v", common.ErrLogical, err)
	}
	scp.Name = scpName.String()
	scp.Network = netName.String()

	s.serviceConnectionPolicies.Add(scp, scpName)

	opName := s.newLongRunningOperationName(scpName.ProjectId())
	b := NewOperationLongRunningBuilder(opName.String(), scpName)
	b.WithDone(true)
	if err := b.WithResult(scp); err != nil {
		return nil, gcpmeta.NewInternalServerError("failed setting create serviceConnectionPolicy operation result: %v", err)
	}
	if err := b.WithNetworkConnectivityMetadata(scpName, "create"); err != nil {
		return nil, fmt.Errorf("%w: failed setting create serviceConnectionPolicy operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*networkconnectivitypb.ServiceConnectionPolicy](b.GetOperationPB()), nil
}

func (s *store) UpdateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.UpdateServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	s.m.Lock()
	s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.ServiceConnectionPolicy == nil {
		return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicy is required")
	}
	if req.UpdateMask == nil || len(req.UpdateMask.Paths) == 0 {
		return nil, gcpmeta.NewBadRequestError("updateMask is required")
	}

	// find existing SCP

	scpName, err := gcputil.ParseNameDetail(req.ServiceConnectionPolicy.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("service connection policy name is invalid: %v", err)
	}
	if scpName.ResourceType() != gcputil.ResourceTypeServiceConnectionPolicy {
		return nil, gcpmeta.NewBadRequestError("service connection policy name type is invalid: %s", scpName.ResourceType())
	}

	scp, err := s.getServiceConnectionPolicyNoLock(scpName.String())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("service connection policy %s does not exist: %v", scpName.String(), err)
	}

	// prevent network change
	if req.ServiceConnectionPolicy.Network != "" && req.ServiceConnectionPolicy.Network != scp.Network {
		return nil, gcpmeta.NewBadRequestError("service connection policy network %s is immutable, wants to update to %s", scp.Network, req.ServiceConnectionPolicy.Network)
	}

	// prevent service class change
	if req.ServiceConnectionPolicy.ServiceClass != "" && req.ServiceConnectionPolicy.ServiceClass != scp.ServiceClass {
		return nil, gcpmeta.NewBadRequestError("service connection policy serviceClass %s is immutable, wants to update to %s", scp.ServiceClass, req.ServiceConnectionPolicy.ServiceClass)
	}

	// validate subnets exist

	for _, subNameTxt := range req.ServiceConnectionPolicy.PscConfig.Subnetworks {
		subName, err := gcputil.ParseNameDetail(subNameTxt)
		if err != nil {
			return nil, gcpmeta.NewBadRequestError("subnetwork name %s is invalid: %v", subNameTxt, err)
		}
		_, err = s.getSubnetNoLock(subName.ProjectId(), subName.LocationRegionId(), subName.ResourceId())
		if err != nil {
			return nil, gcpmeta.NewBadRequestError("subnetwork %s does not exist: %v", subName.String(), err)
		}
	}

	// make the update

	err = UpdateMask(scp, req.ServiceConnectionPolicy, req.UpdateMask)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("service connection policy %s update failed: %v", scpName.String(), err)
	}

	opName := s.newLongRunningOperationName(scpName.ProjectId())
	b := NewOperationLongRunningBuilder(opName.String(), scpName)
	b.WithDone(true)
	if err := b.WithResult(scp); err != nil {
		return nil, gcpmeta.NewInternalServerError("failed setting update serviceConnectionPolicy operation result: %v", err)
	}
	if err := b.WithNetworkConnectivityMetadata(scpName, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting update serviceConnectionPolicy operation metadata: %w", common.ErrLogical, err)
	}
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*networkconnectivitypb.ServiceConnectionPolicy](b.GetOperationPB()), nil
}

func (s *store) GetServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.GetServiceConnectionPolicyRequest, _ ...gax.CallOption) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	s.m.Lock()
	s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicy name is required")
	}

	scp, err := s.getServiceConnectionPolicyNoLock(req.Name)
	if err != nil {
		return nil, err
	}
	return util.Clone(scp)
}

func (s *store) ListServiceConnectionPolicies(ctx context.Context, req *networkconnectivitypb.ListServiceConnectionPoliciesRequest, _ ...gax.CallOption) gcpclient.Iterator[*networkconnectivitypb.ServiceConnectionPolicy] {
	s.m.Lock()
	s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*networkconnectivitypb.ServiceConnectionPolicy]{
			err: ctx.Err(),
		}
	}

	var err error

	list := s.serviceConnectionPolicies
	if req.Parent != "" {
		parentNd, err := gcputil.ParseNameDetail(req.Parent)
		if err != nil {
			return &iteratorMocked[*networkconnectivitypb.ServiceConnectionPolicy]{
				err: gcpmeta.NewBadRequestError("parent service connection policy name is invalid: %v", err),
			}
		}
		list = list.FilterByCallback(func(item FilterableListItem[*networkconnectivitypb.ServiceConnectionPolicy]) bool {
			return item.Name.ProjectId() == parentNd.ProjectId() && item.Name.LocationRegionId() == parentNd.LocationRegionId()
		})
	}
	if req.Filter != "" {
		list, err = list.FilterByExpression(&req.Filter)
		if err != nil {
			return &iteratorMocked[*networkconnectivitypb.ServiceConnectionPolicy]{
				err: gcpmeta.NewBadRequestError("filter expression is invalid: %v", err),
			}
		}
	}

	return list.ToIterator()
}

func (s *store) DeleteServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.DeleteServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("serviceConnectionPolicy name is required")
	}
	scpName, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("service connection policy name is invalid: %v", err)
	}
	if scpName.ResourceType() != gcputil.ResourceTypeServiceConnectionPolicy {
		return nil, gcpmeta.NewBadRequestError("service connection policy name type is invalid: %s", scpName.ResourceType())
	}

	_, err = s.getServiceConnectionPolicyNoLock(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("service connection policy %s does not exist: %v", req.Name, err)
	}

	s.serviceConnectionPolicies = s.serviceConnectionPolicies.FilterNotByCallback(func(item FilterableListItem[*networkconnectivitypb.ServiceConnectionPolicy]) bool {
		return item.Name.Equal(scpName)
	})

	opName := s.newLongRunningOperationName(scpName.ProjectId())
	b := NewOperationLongRunningBuilder(opName.String(), scpName)
	b.WithDone(true)
	if err := b.WithNetworkConnectivityMetadata(scpName, "update"); err != nil {
		return nil, fmt.Errorf("%w: failed setting delete serviceConnectionPolicy operation metadata: %w", common.ErrLogical, err)
	}

	return b.BuildVoidOperation(), nil
}

// high level methods ===========================================================
