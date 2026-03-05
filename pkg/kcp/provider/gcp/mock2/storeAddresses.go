package mock2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

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

/*
address: 10.250.4.0
addressType: INTERNAL
creationTimestamp: '2026-08-30T13:41:08.211-07:00'
description: Some description
id: '7345678901234568267'
kind: compute#address
labelFingerprint: 41XnOqCrTeU=
name: my-address
network: https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network
networkTier: PREMIUM
prefixLength: 22
purpose: VPC_PEERING
selfLink: https://www.googleapis.com/compute/v1/projects/my-project/global/addresses/my-address
status: RESERVED
*/

func (s *store) getAddressNoLock(project, region, address string) (*computepb.Address, error) {
	for _, a := range s.addresses.items {
		if a.Name.ProjectId() == project && a.Name.LocationRegionId() == region && a.Name.ResourceId() == address {
			return a.Obj, nil
		}
	}
	if region == "" {
		return nil, gcpmeta.NewNotFoundError("address %s not found", gcputil.NewGlobalAddressName(project, address).String())
	}
	return nil, gcpmeta.NewNotFoundError("address %s not found", gcputil.NewRegionalAddressName(project, region, address).String())
}

// GlobalAddressesClient methods =======================================================================

func (s *store) GetGlobalAddress(ctx context.Context, req *computepb.GetGlobalAddressRequest, _ ...gax.CallOption) (*computepb.Address, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	addr, err := s.getAddressNoLock(req.Project, "", req.Address)
	if err != nil {
		return nil, err
	}
	cpy, err := copystructure.Copy(addr)
	if err != nil {
		return nil, err
	}
	return cpy.(*computepb.Address), nil
}

func (s *store) DeleteGlobalAddress(ctx context.Context, req *computepb.DeleteGlobalAddressRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	addr, err := s.getAddressNoLock(req.Project, "", req.Address)
	if err != nil {
		return nil, err
	}
	addrName := gcputil.NewGlobalAddressName(req.Project, req.Address)
	if addr.GetAddress() == "" {
		return nil, gcpmeta.NewInternalServerError("%v address %s has empty address", common.ErrLogical, addrName.String())
	}

	networkName, err := gcputil.ParseNameDetail(ptr.Deref(addr.Network, ""))
	if err != nil {
		return nil, fmt.Errorf("%w invalid network reference in address resource: %w", common.ErrLogical, err)
	}
	addressSpace, ok := s.addressSpaces[networkName.String()]
	if !ok {
		return nil, fmt.Errorf("address space not found for address %s in network: %s", addrName.String(), networkName.String())
	}

	// check if address is used

	for _, item := range s.filestores.items {
		for _, nfsNet := range item.Obj.Networks {
			if addrName.EqualString(nfsNet.ReservedIpRange) {
				return nil, gcpmeta.NewBadRequestError("address %s is in use by filestore %s", addrName.String(), item.Name.String())
			}
		}
	}
	for _, item := range s.redisInstances.items {
		if addrName.EqualString(item.Obj.ReservedIpRange) {
			return nil, gcpmeta.NewBadRequestError("address %s is in use by redisInstance %s", addrName.String(), item.Name.String())
		}
	}

	// release the range from address space

	if err := addressSpace.Release(addr.GetAddress()); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v failed to release address space: %v", common.ErrLogical, err)
	}

	s.addresses = s.addresses.FilterNotByCallback(func(item FilterableListItem[*computepb.Address]) bool {
		return item.Name.Equal(addrName)
	})

	op := s.createComputeOperationNoLock(addrName.ProjectId(), "", "delete", addr.GetSelfLink(), addr.GetId())
	op.Status = ptr.To(computepb.Operation_DONE)

	return newComputeOperation(op), nil
}

func (s *store) InsertGlobalAddress(ctx context.Context, req *computepb.InsertGlobalAddressRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Project == "" {
		return nil, gcpmeta.NewBadRequestError("project is required")
	}
	if req.AddressResource == nil {
		return nil, gcpmeta.NewBadRequestError("address resource is required")
	}
	if req.AddressResource.Name == nil {
		return nil, gcpmeta.NewBadRequestError("address name is required")
	}
	if _, err := s.getAddressNoLock(req.Project, "", req.AddressResource.GetName()); err == nil {
		return nil, gcpmeta.NewBadRequestError("address %s already exists", gcputil.NewGlobalAddressName(req.Project, req.AddressResource.GetName()).String())
	}
	if _, ok := computepb.Address_Purpose_value[req.AddressResource.GetPurpose()]; !ok {
		return nil, gcpmeta.NewBadRequestError("invalid address purpose: %q", req.AddressResource.GetPurpose())
	}
	netNd, err := gcputil.ParseNameDetail(ptr.Deref(req.AddressResource.Network, ""))
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid network reference in address resource: %v", err)
	}
	_, err = s.getNetworkNoLock(netNd.ProjectId(), netNd.ResourceId())
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found for address resource", netNd.String())
	}
	if req.AddressResource.AddressType == nil {
		req.AddressResource.AddressType = ptr.To("EXTERNAL")
	}
	if at := ptr.Deref(req.AddressResource.AddressType, ""); at != "EXTERNAL" && at != "INTERNAL" {
		return nil, gcpmeta.NewBadRequestError("invalid address type: %q", at)
	}

	// create address

	addr, err := util.Clone(req.AddressResource)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("failed to clone address resource: %v", err)
	}

	// add cidr to address space

	addressSpace, ok := s.addressSpaces[netNd.String()]
	if !ok {
		return nil, gcpmeta.NewInternalServerError("%v address space for network %q not found", common.ErrLogical, netNd.String())
	}

	c := fmt.Sprintf("%s/%d", addr.GetAddress(), addr.GetPrefixLength())
	if err := addressSpace.Reserve(c); err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid address range %q: %v", c, err)
	}

	id := rand.Uint64()
	name := gcputil.NewGlobalAddressName(req.Project, addr.GetName())
	addr.Id = ptr.To(id)
	addr.SelfLink = ptr.To(name.PrefixWithGoogleApisComputeV1())
	addr.Kind = ptr.To("compute#address")
	addr.Status = ptr.To(computepb.Address_RESERVED.String())

	s.addresses.Add(addr, name)

	op := s.createComputeOperationNoLock(name.ProjectId(), "", "insert", addr.GetSelfLink(), addr.GetId())
	op.Status = ptr.To(computepb.Operation_DONE)

	return newComputeOperation(op), nil
}

func (s *store) ListGlobalAddresses(ctx context.Context, req *computepb.ListGlobalAddressesRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Address] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Address]{err: ctx.Err()}
	}

	list := s.addresses
	if req.Project != "" {
		list = s.addresses.FilterByCallback(func(l FilterableListItem[*computepb.Address]) bool {
			return l.Name.ProjectId() == req.Project && l.Name.LocationRegionId() == ""
		})
	}
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Address]{err: gcpmeta.NewBadRequestError("invalid filter: %v", err)}
	}

	return list.ToIterator()
}

// RegionalAddressesClient methods =======================================================================

func (s *store) ListAddresses(ctx context.Context, req *computepb.ListAddressesRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Address] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*computepb.Address]{err: ctx.Err()}
	}

	list := s.addresses
	if req.Project != "" {
		list = s.addresses.FilterByCallback(func(l FilterableListItem[*computepb.Address]) bool {
			return l.Name.ProjectId() == req.Project
		})
	}
	if req.Region != "" {
		list = s.addresses.FilterByCallback(func(l FilterableListItem[*computepb.Address]) bool {
			return l.Name.LocationRegionId() == req.Region
		})
	} else {
		list = s.addresses.FilterByCallback(func(l FilterableListItem[*computepb.Address]) bool {
			return l.Name.LocationRegionId() != ""
		})
	}
	var err error
	list, err = list.FilterByExpression(req.Filter)
	if err != nil {
		return &iteratorMocked[*computepb.Address]{err: gcpmeta.NewBadRequestError("invalid filter: %v", err)}
	}

	return list.ToIterator()
}

// Higher level RegionalAddressesClient functions --------------------------------------------------------

func (s *store) GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error) {
	it := s.ListAddresses(ctx, &computepb.ListAddressesRequest{
		Project: project,
		Region:  region,
		Filter:  ptr.To(`purpose="NAT_AUTO"`), // the API does not work with users filter, so have to do this
	}).All()
	var results []*computepb.Address
	for x, err := range it {
		if err != nil {
			return nil, err
		}
		for _, usr := range x.Users {
			if strings.HasSuffix(usr, "/"+routerName) {
				results = append(results, x)
				break
			}
		}
	}
	return results, nil
}
