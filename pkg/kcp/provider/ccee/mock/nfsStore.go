package mock

import (
	"context"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"net/http"
	"sync"
	"time"
)

func deptrSlice[T any](arr []*T) []T {
	return pie.Map(arr, func(t *T) T {
		return *t
	})
}

type NfsConfig interface {
	AddNetwork(id, name string) *networks.Network
	AddSubnet(id, networkId, name, cidr string) *subnets.Subnet
	SetShareStatus(id, status string)
}

type nfsStore struct {
	m             sync.Mutex
	networks      []*networks.Network
	subnets       map[string][]*subnets.Subnet
	shareNetworks map[string][]*sharenetworks.ShareNetwork
	shares        map[string][]*shares.Share
	access        map[string][]*shares.AccessRight
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

func (s *nfsStore) ListShares(ctx context.Context, shareNetworkId string) ([]shares.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*shares.Share{}
	}
	arr, ok := s.shares[shareNetworkId]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GetShare(ctx context.Context, id string) (*shares.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*shares.Share{}
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

func (s *nfsStore) CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.shares == nil {
		s.shares = map[string][]*shares.Share{}
	}
	sh := &shares.Share{
		ID:             uuid.NewString(),
		ShareNetworkID: shareNetworkId,
		Name:           name,
		Size:           size,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
		Status:         "creating",
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
		s.shares = map[string][]*shares.Share{}
	}
	for netId, arr := range s.shares {
		s.shares[netId] = pie.Filter(arr, func(x *shares.Share) bool {
			return x.ID != id
		})
	}
	return nil
}

func (s *nfsStore) ListShareExportLocations(ctx context.Context, id string) ([]shares.ExportLocation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	time.Sleep(time.Millisecond)
	return []shares.ExportLocation{
		{
			Path:      fmt.Sprintf("10.100.0.10:/%s-1", id),
			Preferred: true,
		},
		{
			Path:      fmt.Sprintf("10.200.0.20:/%s-2", id),
			Preferred: false,
		},
	}, nil
}

func (s *nfsStore) ListShareAccessRights(ctx context.Context, id string) ([]shares.AccessRight, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.access == nil {
		s.access = map[string][]*shares.AccessRight{}
	}
	arr, ok := s.access[id]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *nfsStore) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*shares.AccessRight, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	if s.access == nil {
		s.access = map[string][]*shares.AccessRight{}
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
	a := &shares.AccessRight{
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
		s.access = map[string][]*shares.AccessRight{}
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
	arr = pie.Filter(arr, func(x *shares.AccessRight) bool {
		return x.ID != accessId
	})
	s.access[shareId] = arr
	return nil
}
