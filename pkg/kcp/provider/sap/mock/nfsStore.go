package mock

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
)

func deptrSlice[T any](arr []*T) []T {
	return pie.Map(arr, func(t *T) T {
		return *t
	})
}

type NfsConfig interface {
	AddNetwork(id, name string) *networks.Network
	AddRouter(id, name string, ipAddresses ...string) *routers.Router
	AddSubnet(id, networkId, name, cidr string) *subnets.Subnet
	SetShareStatus(id, status string)
}

type nfsStore struct {
	m             sync.Mutex
	networks      []*networks.Network
	routers       []*routers.Router
	subnets       map[string][]*subnets.Subnet
	shareNetworks map[string][]*sharenetworks.ShareNetwork
	shares        map[string][]*sapnfsinstanceclient.Share
	access        map[string][]*sapnfsinstanceclient.ShareAccess
}

// NfsConfig implementation ======================================================================

func (s *nfsStore) AddNetwork(id, name string) *networks.Network {
	s.m.Lock()
	defer s.m.Unlock()

	n := &networks.Network{
		ID:     id,
		Name:   name,
		Status: "ACTIVE",
	}
	s.networks = append(s.networks, n)
	return n
}

func (s *nfsStore) AddRouter(id, name string, ipAddresses ...string) *routers.Router {
	s.m.Lock()
	defer s.m.Unlock()

	r := &routers.Router{
		ID:   id,
		Name: name,
	}
	subnetId := uuid.NewString()
	for _, ip := range ipAddresses {
		r.GatewayInfo.ExternalFixedIPs = append(r.GatewayInfo.ExternalFixedIPs, routers.ExternalFixedIP{
			IPAddress: ip,
			SubnetID:  subnetId,
		})
	}

	s.routers = append(s.routers, r)
	return r
}

func (s *nfsStore) AddSubnet(id, networkId, name, cidr string) *subnets.Subnet {
	s.m.Lock()
	defer s.m.Unlock()

	var network *networks.Network
	for _, n := range s.networks {
		if n.ID == networkId {
			network = n
			break
		}
	}
	if network != nil {
		network.Subnets = pie.Unique(append(network.Subnets, id))
	}

	if s.subnets == nil {
		s.subnets = map[string][]*subnets.Subnet{}
	}

	subnet := &subnets.Subnet{
		ID:        id,
		NetworkID: networkId,
		Name:      name,
		IPVersion: 4,
		CIDR:      cidr,
	}

	s.subnets[networkId] = append(s.subnets[networkId], subnet)

	return subnet
}

func (s *nfsStore) SetShareStatus(id, status string) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, arr := range s.shares {
		for _, share := range arr {
			if share.ID == id {
				share.Status = status
				return
			}
		}
	}
}

// NfsInstanceClient implementation ==============================================================

func (s *nfsStore) ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	var result []networks.Network
	for _, network := range s.networks {
		if name == "" || network.Name == name {
			result = append(result, *network)
		}
	}
	return result, nil
}

func (s *nfsStore) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	for _, network := range s.networks {
		if network.ID == id {
			return network, nil
		}
	}
	return nil, nil
}

func (s *nfsStore) GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	for _, router := range s.routers {
		if router.Name == routerName {
			return router, nil
		}
	}
	return nil, nil
}

func (s *nfsStore) ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.subnets == nil {
		s.subnets = map[string][]*subnets.Subnet{}
	}
	arr, ok := s.subnets[networkId]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.subnets == nil {
		s.subnets = map[string][]*subnets.Subnet{}
	}
	for _, arr := range s.subnets {
		for _, subnet := range arr {
			if subnet.ID == id {
				return subnet, nil
			}
		}
	}
	return nil, nil
}

func (s *nfsStore) ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shareNetworks == nil {
		s.shareNetworks = map[string][]*sharenetworks.ShareNetwork{}
	}
	arr, ok := s.shareNetworks[networkId]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error) {
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shareNetworks == nil {
		s.shareNetworks = map[string][]*sharenetworks.ShareNetwork{}
	}
	for _, arr := range s.shareNetworks {
		for _, sn := range arr {
			if sn.ID == id {
				return sn, nil
			}
		}
	}
	return nil, nil
}

func (s *nfsStore) CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shareNetworks == nil {
		s.shareNetworks = map[string][]*sharenetworks.ShareNetwork{}
	}
	sn := &sharenetworks.ShareNetwork{
		ID:              uuid.NewString(),
		NeutronNetID:    networkId,
		NeutronSubnetID: subnetId,
		Name:            name,
	}
	s.shareNetworks[networkId] = append(s.shareNetworks[networkId], sn)
	return sn, nil
}

func (s *nfsStore) DeleteShareNetwork(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	if s.shares != nil {
		for _, arr := range s.shares {
			for _, sn := range arr {
				if sn.ShareNetworkID == id {
					return errors.New("can not delete share network with existing shares")
				}
			}
		}
	}

	if s.shareNetworks == nil {
		s.shareNetworks = map[string][]*sharenetworks.ShareNetwork{}
	}
	for netId, arr := range s.shareNetworks {
		s.shareNetworks[netId] = pie.Filter(arr, func(x *sharenetworks.ShareNetwork) bool {
			return x.ID != id
		})
	}

	return nil
}

func (s *nfsStore) ListShares(ctx context.Context, shareNetworkId string) ([]sapnfsinstanceclient.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*sapnfsinstanceclient.Share{}
	}
	arr, ok := s.shares[shareNetworkId]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GetShare(ctx context.Context, id string) (*sapnfsinstanceclient.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*sapnfsinstanceclient.Share{}
	}
	for _, arr := range s.shares {
		for _, share := range arr {
			if share.ID == id {
				return share, nil
			}
		}
	}
	return nil, nil
}

func (s *nfsStore) CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*sapnfsinstanceclient.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*sapnfsinstanceclient.Share{}
	}
	id := uuid.NewString()
	sh := &sapnfsinstanceclient.Share{
		ID:             id,
		ShareNetworkID: shareNetworkId,
		Name:           name,
		Size:           size,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
		Status:         "creating",
		ExportLocation: fmt.Sprintf("10.100.0.10:/%s-1", id),
		ExportLocations: []string{
			fmt.Sprintf("10.100.0.10:/%s-1", id),
			fmt.Sprintf("10.100.0.10:/%s-2", id),
		},
	}
	s.shares[shareNetworkId] = append(s.shares[shareNetworkId], sh)
	return sh, nil
}

func (s *nfsStore) DeleteShare(ctx context.Context, id string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*sapnfsinstanceclient.Share{}
	}
	for netId, arr := range s.shares {
		s.shares[netId] = pie.Filter(arr, func(x *sapnfsinstanceclient.Share) bool {
			return x.ID != id
		})
	}
	return nil
}

func (s *nfsStore) ShareShrink(ctx context.Context, shareId string, newSize int) error {
	return s.shareChangeSize(ctx, shareId, newSize)
}

func (s *nfsStore) ShareExtend(ctx context.Context, shareId string, newSize int) error {
	return s.shareChangeSize(ctx, shareId, newSize)
}

func (s *nfsStore) shareChangeSize(ctx context.Context, shareId string, newSize int) error {
	s.m.Lock()
	defer s.m.Unlock()
	var theShare *sapnfsinstanceclient.Share
	for _, arr := range s.shares {
		for _, sh := range arr {
			if sh.ID == shareId {
				theShare = sh
			}
		}
	}
	if theShare == nil {
		return &gophercloud.ErrUnexpectedResponseCode{
			BaseError: gophercloud.BaseError{
				Info: fmt.Sprintf("share %q does not exist", shareId),
			},
			Actual: http.StatusNotFound,
		}
	}
	theShare.Size = newSize
	return nil
}

func (s *nfsStore) ListShareAccessRules(ctx context.Context, id string) ([]sapnfsinstanceclient.ShareAccess, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.access == nil {
		s.access = map[string][]*sapnfsinstanceclient.ShareAccess{}
	}
	arr, ok := s.access[id]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*sapnfsinstanceclient.ShareAccess, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.access == nil {
		s.access = map[string][]*sapnfsinstanceclient.ShareAccess{}
	}

	exists := false
	for _, arr := range s.shares {
		for _, share := range arr {
			if share.ID == shareId {
				exists = true
				break
			}
		}
	}
	if !exists {
		return nil, &gophercloud.ErrUnexpectedResponseCode{
			BaseError: gophercloud.BaseError{
				Info: fmt.Sprintf("share %q does not exist", shareId),
			},
			Actual: http.StatusNotFound,
		}
	}

	arr := s.access[shareId]
	a := &sapnfsinstanceclient.ShareAccess{
		ID:          uuid.NewString(),
		ShareID:     shareId,
		AccessType:  "ip",
		AccessTo:    cidr,
		AccessLevel: "rw",
	}
	arr = append(arr, a)
	s.access[shareId] = arr
	return a, nil
}

func (s *nfsStore) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.access == nil {
		s.access = map[string][]*sapnfsinstanceclient.ShareAccess{}
	}

	exists := false
	for _, arr := range s.shares {
		for _, share := range arr {
			if share.ID == shareId {
				exists = true
				break
			}
		}
	}
	if !exists {
		return &gophercloud.ErrUnexpectedResponseCode{
			BaseError: gophercloud.BaseError{
				Info: fmt.Sprintf("share %q does not exist", shareId),
			},
			Actual: http.StatusNotFound,
		}
	}

	arr := s.access[shareId]
	arr = pie.Filter(arr, func(x *sapnfsinstanceclient.ShareAccess) bool {
		return x.ID != accessId
	})
	s.access[shareId] = arr
	return nil
}
