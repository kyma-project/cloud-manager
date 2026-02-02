package mock

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deptrSlice[T any](arr []*T) []T {
	return pie.Map(arr, func(t *T) T {
		return *t
	})
}

type NfsConfig interface {
	AddNetwork(id, name string) *networks.Network
	AddRouter(id, name string, ipAddresses ...string) *routers.Router
	SetShareStatus(id, status string)
}

func newMainStore() *mainStore {
	return &mainStore{
		subnets:       map[string][]subnetInfoType{},
		shareNetworks: map[string][]*sharenetworks.ShareNetwork{},
		shares:        map[string][]*shares.Share{},
		access:        map[string][]*sapclient.ShareAccess{},
	}
}

type subnetInfoType struct {
	subnet       *subnets.Subnet
	addressSpace allocate.AddressSpace
}

type mainStore struct {
	m                sync.Mutex
	internalNetworks []*networks.Network
	externalNetworks []*networks.Network
	routers          []*routers.Router
	ports            []*ports.Port
	subnets          map[string][]subnetInfoType
	shareNetworks    map[string][]*sharenetworks.ShareNetwork
	shares           map[string][]*shares.Share
	access           map[string][]*sapclient.ShareAccess
}

// NfsConfig implementation ======================================================================

func (s *mainStore) AddNetwork(id, name string) *networks.Network {
	s.m.Lock()
	defer s.m.Unlock()

	n := &networks.Network{
		ID:     id,
		Name:   name,
		Status: networks.StatusActive,
	}
	s.internalNetworks = append(s.internalNetworks, n)
	return n
}

func (s *mainStore) AddRouter(id, name string, ipAddresses ...string) *routers.Router {
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

func (s *mainStore) SetShareStatus(id, status string) {
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

// Clients implementation ==============================================================

var _ Clients = (*mainStore)(nil)

// NetworkClient implementation ------------------------------------------------

func (s *mainStore) ListNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error) {
	return s.listGivenNetworks(ctx, opts, s.internalNetworks)
}

func (s *mainStore) ListExternalNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error) {
	return s.listGivenNetworks(ctx, opts, s.externalNetworks)
}

func (s *mainStore) listGivenNetworks(ctx context.Context, opts networks.ListOpts, givenNetworks []*networks.Network) ([]networks.Network, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	var result []networks.Network
	for _, network := range givenNetworks {
		isMatch := true
		if opts.Name != "" && network.Name != opts.Name {
			isMatch = false
		}
		if opts.ID != "" && network.ID != opts.ID {
			isMatch = false
		}
		if opts.AdminStateUp != nil && network.AdminStateUp != *opts.AdminStateUp {
			isMatch = false
		}
		if opts.Status != "" && network.Status != opts.Status {
			isMatch = false
		}
		if isMatch {
			result = append(result, *network)
		}
	}
	return result, nil
}

func (s *mainStore) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	for _, network := range s.internalNetworks {
		if network.ID == id {
			return network, nil
		}
	}
	for _, network := range s.externalNetworks {
		if network.ID == id {
			return network, nil
		}
	}
	return nil, nil
}

func (s *mainStore) CreateExternalNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error) {
	return s.CreateNetwork(ctx, external.CreateOptsExt{
		CreateOptsBuilder: opts,
		External:          ptr.To(true),
	})
}

func (s *mainStore) CreateNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error) {
	s.m.Lock()
	defer s.m.Unlock()

	isExternal := false
	var optsInternal networks.CreateOpts
	switch oo := opts.(type) {
	case external.CreateOptsExt:
		isExternal = true
		optsInternal = oo.CreateOptsBuilder.(networks.CreateOpts)
	case networks.CreateOpts:
		optsInternal = oo
	default:
		return nil, fmt.Errorf("invalid options type: %T", opts)
	}

	if optsInternal.Name == "" {
		return nil, fmt.Errorf("openstack network name is required")
	}

	up := true
	if optsInternal.AdminStateUp != nil {
		up = *optsInternal.AdminStateUp
	}

	n := &networks.Network{
		ID:           uuid.NewString(),
		Name:         optsInternal.Name,
		Description:  optsInternal.Description,
		Status:       networks.StatusActive,
		AdminStateUp: up,
	}

	if isExternal {
		s.externalNetworks = append(s.externalNetworks, n)
	} else {
		s.internalNetworks = append(s.internalNetworks, n)
	}

	return n, nil
}

func (s *mainStore) DeleteNetwork(ctx context.Context, id string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	isFound := false
	for _, n := range s.internalNetworks {
		if n.ID == id {
			isFound = true
			break
		}
	}
	isExternal := false
	if !isFound {
		for _, n := range s.externalNetworks {
			if n.ID == id {
				isExternal = true
				isFound = true
				break
			}
		}
	}

	if !isFound {
		return sapmeta.NewNotFoundError("Not found")
	}

	if isExternal {
		for _, router := range s.routers {
			if router.GatewayInfo.NetworkID == id {
				return sapmeta.NewBadRequestError("router using this external network")
			}
		}
	} else {
		_, exists := s.subnets[id]
		if exists && len(s.subnets[id]) > 0 {
			return sapmeta.NewBadRequestError("network has existing subnets")
		}

		_, exists = s.shareNetworks[id]
		if exists && len(s.shareNetworks[id]) > 0 {
			return sapmeta.NewBadRequestError("network has existing share networks")
		}
	}

	if isExternal {
		s.externalNetworks = pie.FilterNot(s.externalNetworks, func(n *networks.Network) bool {
			return n.ID == id
		})
	} else {
		s.internalNetworks = pie.FilterNot(s.internalNetworks, func(n *networks.Network) bool {
			return n.ID == id
		})
	}
	return nil
}

// NetworkClient high level derived methods --------------------------------------------

func (s *mainStore) ListInternalNetworksByName(ctx context.Context, name string) ([]networks.Network, error) {
	return s.ListNetworks(ctx, networks.ListOpts{Name: name})
}

func (s *mainStore) GetNetworkByName(ctx context.Context, networkName string) (*networks.Network, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	for _, network := range s.internalNetworks {
		if network.Name == networkName {
			return network, nil
		}
	}
	return nil, nil
}

// SubnetClient implementation -------------------------------------------------

func (s *mainStore) ListSubnets(ctx context.Context, opts subnets.ListOpts) ([]subnets.Subnet, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	var result []subnets.Subnet
	for _, arrSubnets := range s.subnets {
		for _, subnet := range arrSubnets {
			isMatch := true
			if opts.ID != "" && subnet.subnet.ID != opts.ID {
				isMatch = false
			}
			if opts.NetworkID != "" && subnet.subnet.NetworkID != opts.NetworkID {
				isMatch = false
			}
			if opts.Name != "" && subnet.subnet.Name != opts.Name {
				isMatch = false
			}
			if opts.GatewayIP != "" && subnet.subnet.GatewayIP != opts.GatewayIP {
				isMatch = false
			}
			if isMatch {
				result = append(result, *subnet.subnet)
			}
		}
	}
	return result, nil
}

func (s *mainStore) getSubnetInfoByIdNoLock(id string) subnetInfoType {
	for _, arr := range s.subnets {
		for _, subnet := range arr {
			if subnet.subnet.ID == id {
				return subnet
			}
		}
	}
	return subnetInfoType{}
}

func (s *mainStore) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	return s.getSubnetInfoByIdNoLock(id).subnet, nil
}

func (s *mainStore) CreateSubnet(ctx context.Context, opts subnets.CreateOpts) (*subnets.Subnet, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	var network *networks.Network
	//isExternal := false
	for _, net := range s.internalNetworks {
		if net.ID == opts.NetworkID {
			network = net
			break
		}
	}
	if network == nil {
		for _, net := range s.externalNetworks {
			if net.ID == opts.NetworkID {
				network = net
				//isExternal = true
				break
			}
		}
	}
	if network == nil {
		return nil, sapmeta.NewNotFoundError("network not found")
	}

	subnetId := uuid.NewString()

	network.Subnets = pie.Unique(append(network.Subnets, subnetId))

	as, err := allocate.NewAddressSpace(opts.CIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR for subnet %q: %w", opts.CIDR, err)
	}

	subnet := &subnets.Subnet{
		ID:          subnetId,
		NetworkID:   opts.NetworkID,
		Name:        opts.Name,
		Description: opts.Description,
		IPVersion:   int(opts.IPVersion),
		CIDR:        opts.CIDR,
	}

	s.subnets[network.ID] = append(s.subnets[network.ID], subnetInfoType{
		subnet:       subnet,
		addressSpace: as,
	})

	return subnet, nil
}

func (s *mainStore) DeleteSubnet(ctx context.Context, subnetId string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	foundInNetworkId := ""

	for networkId, arrSubnets := range s.subnets {
		for _, subnet := range arrSubnets {
			if subnet.subnet.ID == subnetId {
				foundInNetworkId = networkId
				break
			}
		}
	}

	if foundInNetworkId == "" {
		return &gophercloud.ErrUnexpectedResponseCode{
			BaseError: gophercloud.BaseError{
				Info: fmt.Sprintf("subnet %q not found", subnetId),
			},
			Expected: []int{http.StatusOK},
			Actual:   http.StatusNotFound,
		}
	}

	for _, net := range s.internalNetworks {
		if net.ID == foundInNetworkId {
			net.Subnets = pie.FilterNot(net.Subnets, func(x string) bool {
				return x == subnetId
			})
			break
		}
	}

	s.subnets[foundInNetworkId] = pie.FilterNot(s.subnets[foundInNetworkId], func(subnetInfo subnetInfoType) bool {
		return subnetInfo.subnet.ID == subnetId
	})

	return nil
}

// SubnetClient high level derived methods --------------------------------------------

func (s *mainStore) GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	for _, arr := range s.subnets {
		for _, subnet := range arr {
			if subnet.subnet.NetworkID == networkId && subnet.subnet.Name == subnetName {
				return subnet.subnet, nil
			}
		}
	}
	return nil, nil
}

func (s *mainStore) CreateSubnetOp(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error) {
	return s.CreateSubnet(ctx, subnets.CreateOpts{
		NetworkID: networkId,
		CIDR:      cidr,
		Name:      subnetName,
	})
}

func (s *mainStore) ListSubnetsByNetworkId(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	return s.ListSubnets(ctx, subnets.ListOpts{NetworkID: networkId})
}

// RouterClient implementation -------------------------------------------------

func (s *mainStore) ListRouters(ctx context.Context, opts routers.ListOpts) ([]routers.Router, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	var result []routers.Router
	for _, router := range s.routers {
		isMatch := true
		if opts.Name != "" && router.Name != opts.Name {
			isMatch = false
		}
		if opts.ID != "" && router.ID != opts.ID {
			isMatch = false
		}
		if opts.AdminStateUp != nil && *opts.AdminStateUp != router.AdminStateUp {
			isMatch = false
		}
		if opts.Status != "" && router.Status != opts.Status {
			isMatch = false
		}
		if isMatch {
			result = append(result, *router)
		}
	}
	return result, nil
}

func (s *mainStore) GetRouter(ctx context.Context, id string) (*routers.Router, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	return s.getRouterNoLock(ctx, id)
}

func (s *mainStore) getRouterNoLock(ctx context.Context, id string) (*routers.Router, error) {
	for _, router := range s.routers {
		if router.ID == id {
			return router, nil
		}
	}
	return nil, nil
}

func (s *mainStore) CreateRouter(ctx context.Context, opts routers.CreateOpts) (*routers.Router, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	theRouter := &routers.Router{
		ID:          uuid.NewString(),
		Name:        opts.Name,
		Description: opts.Description,
		Status:      "ACTIVE",
	}
	if opts.GatewayInfo != nil {
		theRouter.GatewayInfo.NetworkID = opts.GatewayInfo.NetworkID
		theRouter.GatewayInfo.ExternalFixedIPs = append([]routers.ExternalFixedIP{}, opts.GatewayInfo.ExternalFixedIPs...)
		for i, ip := range theRouter.GatewayInfo.ExternalFixedIPs {
			if ip.IPAddress != "" {
				continue
			}
			subnetInfo := s.getSubnetInfoByIdNoLock(ip.SubnetID)
			if subnetInfo.subnet == nil {
				return nil, sapmeta.NewNotFoundError("subnet not found")
			}
			ipAddr, err := subnetInfo.addressSpace.AllocateOneIpAddress()
			if err != nil {
				return nil, fmt.Errorf("error allocating IP address for router subnet: %w", err)
			}
			theRouter.GatewayInfo.ExternalFixedIPs[i].IPAddress = ipAddr
		}
	}
	s.routers = append(s.routers, theRouter)
	return theRouter, nil
}

func (s *mainStore) UpdateRouter(ctx context.Context, routerId string, opts routers.UpdateOpts) (*routers.Router, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	theRouter, err := s.getRouterNoLock(ctx, routerId)
	if err != nil {
		return nil, fmt.Errorf("error getting router: %w", err)
	}
	if theRouter == nil {
		return nil, sapmeta.NewNotFoundError("router not found")
	}

	theRouter.Name = opts.Name
	if opts.Description != nil {
		theRouter.Description = *opts.Description
	}
	if opts.GatewayInfo != nil {
		theRouter.GatewayInfo.NetworkID = opts.GatewayInfo.NetworkID
		theRouter.GatewayInfo.ExternalFixedIPs = append([]routers.ExternalFixedIP{}, opts.GatewayInfo.ExternalFixedIPs...)
	}
	return theRouter, nil
}

func (s *mainStore) DeleteRouter(ctx context.Context, id string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	isFound := false
	s.routers = pie.FilterNot(s.routers, func(r *routers.Router) bool {
		if r.ID == id {
			isFound = true
			return true
		}
		return false
	})
	if !isFound {
		return sapmeta.NewNotFoundError("router not found")
	}
	return nil
}

func (s *mainStore) AddRouterInterface(ctx context.Context, routerId string, opts routers.AddInterfaceOpts) (*routers.InterfaceInfo, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	theRouter, err := s.getRouterNoLock(ctx, routerId)
	if err != nil {
		return nil, fmt.Errorf("error getting router: %w", err)
	}
	if theRouter == nil {
		return nil, sapmeta.NewNotFoundError("router not found")
	}

	subnetInfo := s.getSubnetInfoByIdNoLock(opts.SubnetID)
	if subnetInfo.subnet == nil {
		return nil, sapmeta.NewNotFoundError("subnet not found")
	}
	for _, ext := range theRouter.GatewayInfo.ExternalFixedIPs {
		if ext.SubnetID == opts.SubnetID {
			return nil, sapmeta.NewBadRequestError("subnet already added to router")
		}
	}

	arrPorts, err := s.listPortsNoLock(ctx, ports.ListOpts{
		DeviceID: routerId,
		FixedIPs: []ports.FixedIPOpts{
			{SubnetID: subnetInfo.subnet.ID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing router ports: %w", err)
	}
	if len(arrPorts) > 0 {
		return nil, sapmeta.NewBadRequestError("subnet already added to router")
	}

	var port *ports.Port
	if opts.PortID == "" {
		p, err := s.createPortNoLock(ctx, ports.CreateOpts{
			DeviceID:     routerId,
			NetworkID:    subnetInfo.subnet.NetworkID,
			Name:         uuid.NewString(),
			AdminStateUp: ptr.To(true),
			FixedIPs: []ports.IP{{
				SubnetID: opts.SubnetID,
			}},
		})
		if err != nil {
			return nil, fmt.Errorf("error creating new port for router interface: %w", err)
		}
		port = p
	} else {
		return nil, sapmeta.NewBadRequestError("add router interface with existing port not supported in mock")
	}

	return &routers.InterfaceInfo{
		SubnetID: opts.SubnetID,
		PortID:   port.ID,
		ID:       port.ID,
	}, nil
}

func (s *mainStore) RemoveRouterInterface(ctx context.Context, routerId string, opts routers.RemoveInterfaceOpts) (*routers.InterfaceInfo, error) {
	if opts.SubnetID == "" {
		return nil, sapmeta.NewBadRequestError("subnet ID is required to remove router interface in mock")
	}

	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	router, err := s.getRouterNoLock(ctx, routerId)
	if err != nil {
		return nil, fmt.Errorf("error getting router: %w", err)
	}
	if router == nil {
		return nil, sapmeta.NewNotFoundError("router not found")
	}

	arrPorts, err := s.listPortsNoLock(ctx, ports.ListOpts{
		DeviceID: routerId,
		FixedIPs: []ports.FixedIPOpts{
			{SubnetID: opts.SubnetID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing router ports: %w", err)
	}
	if len(arrPorts) == 0 {
		return nil, sapmeta.NewBadRequestError("router port not found in the given subnet")
	}

	for _, port := range arrPorts {
		err := s.deletePortNoLock(ctx, port.ID)
		if err != nil {
			return nil, fmt.Errorf("error deleting port for router subnet interface: %w", err)
		}
	}

	return &routers.InterfaceInfo{
		SubnetID: opts.SubnetID,
		PortID:   arrPorts[0].ID,
		ID:       arrPorts[0].ID,
	}, nil
}

// RouterClient high level derived methods --------------------------------------------

func (s *mainStore) GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error) {
	if util.IsContextDone(ctx) {
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

func (s *mainStore) AddSubnetToRouter(ctx context.Context, routerId string, subnetId string) (*routers.InterfaceInfo, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	subnetInfo := s.getSubnetInfoByIdNoLock(subnetId)
	if subnetInfo.subnet == nil {
		return nil, sapclient.NewNotFoundError(fmt.Sprintf("subnet %q not found", subnetId))
	}

	arrPorts, err := s.listPortsNoLock(ctx, ports.ListOpts{
		DeviceID: routerId,
		FixedIPs: []ports.FixedIPOpts{
			{SubnetID: subnetId},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing router ports: %w", err)
	}
	if len(arrPorts) > 0 {
		return nil, sapmeta.NewBadRequestError("subnet already added to router")
	}

	port, err := s.createPortNoLock(ctx, ports.CreateOpts{
		DeviceID:     routerId,
		NetworkID:    subnetInfo.subnet.NetworkID,
		Name:         fmt.Sprintf("%s-%s", routerId, subnetId),
		AdminStateUp: ptr.To(true),
		FixedIPs: []ports.IP{{
			SubnetID: subnetId,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating new port for router subnet interface: %w", err)
	}

	return &routers.InterfaceInfo{
		SubnetID: subnetId,
		PortID:   port.ID,
		ID:       port.ID,
	}, nil
}

func (s *mainStore) RemoveSubnetFromRouter(ctx context.Context, routerId string, subnetId string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	_, err := s.getRouterNoLock(ctx, routerId)
	if err != nil {
		return sapmeta.NewNotFoundError("router not found")
	}
	subnetInfo := s.getSubnetInfoByIdNoLock(subnetId)
	if subnetInfo.subnet == nil {
		return sapmeta.NewNotFoundError("subnet not found")
	}

	arrPorts, err := s.listPortsNoLock(ctx, ports.ListOpts{
		DeviceID: routerId,
		FixedIPs: []ports.FixedIPOpts{
			{SubnetID: subnetId},
		},
	})
	if err != nil {
		return fmt.Errorf("error listing router ports: %w", err)
	}

	for _, port := range arrPorts {
		err := s.deletePortNoLock(ctx, port.ID)
		if err != nil {
			return fmt.Errorf("error deleting port for router subnet interface: %w", err)
		}
	}

	return nil
}

// PortClient implementation -------------------------------------------------

func (s *mainStore) ListPorts(ctx context.Context, opts ports.ListOpts) ([]ports.Port, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	return s.listPortsNoLock(ctx, opts)
}

func (s *mainStore) listPortsNoLock(ctx context.Context, opts ports.ListOpts) ([]ports.Port, error) {
	var result []ports.Port
	for _, port := range s.ports {
		isMatch := true
		if opts.ID != "" && port.ID != opts.ID {
			isMatch = false
		}
		if opts.DeviceID != "" && port.DeviceID != opts.DeviceID {
			isMatch = false
		}
		if opts.Name != "" && port.Name != opts.Name {
			isMatch = false
		}
		if opts.NetworkID != "" && port.NetworkID != opts.NetworkID {
			isMatch = false
		}
		if opts.Status != "" && port.Status != opts.Status {
			isMatch = false
		}
		if opts.AdminStateUp != nil && port.AdminStateUp != *opts.AdminStateUp {
			isMatch = false
		}
		for _, fip := range opts.FixedIPs {
			if fip.SubnetID != "" {
				found := false
				for _, portFip := range port.FixedIPs {
					if portFip.SubnetID == fip.SubnetID {
						found = true
						break
					}
				}
				if !found {
					isMatch = false
				}
			}
			if fip.IPAddress != "" {
				found := false
				for _, portFip := range port.FixedIPs {
					if portFip.IPAddress == fip.IPAddress {
						found = true
						break
					}
				}
				if !found {
					isMatch = false
				}
			}
		}
		if isMatch {
			result = append(result, *port)
		}
	}
	return result, nil
}

func (s *mainStore) GetPort(ctx context.Context, id string) (*ports.Port, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	return s.getPortNoLock(ctx, id)
}

func (s *mainStore) getPortNoLock(ctx context.Context, id string) (*ports.Port, error) {
	for _, port := range s.ports {
		if port.ID == id {
			return port, nil
		}
	}
	return nil, nil
}

func (s *mainStore) CreatePort(ctx context.Context, opts ports.CreateOpts) (*ports.Port, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	return s.createPortNoLock(ctx, opts)
}

func (s *mainStore) createPortNoLock(ctx context.Context, opts ports.CreateOpts) (*ports.Port, error) {
	ips, ok := opts.FixedIPs.([]ports.IP)
	if !ok {
		return nil, sapmeta.NewBadRequestError("FixedIPs is not a array of IP address")
	}
	if len(ips) == 0 {
		return nil, sapmeta.NewBadRequestError("FixedIPs is empty")
	}
	if opts.NetworkID == "" {
		return nil, sapmeta.NewBadRequestError("NetworkID is empty")
	}

	var fips []ports.IP
	for _, ip := range ips {
		subnetInfo := s.getSubnetInfoByIdNoLock(ip.SubnetID)
		if subnetInfo.subnet == nil {
			return nil, sapmeta.NewNotFoundError("Subnet not found")
		}
		if opts.NetworkID != "" && opts.NetworkID != subnetInfo.subnet.NetworkID {
			return nil, sapmeta.NewBadRequestError("Subnet does not belong to given port NetworkID")
		}

		addr, err := subnetInfo.addressSpace.AllocateOneIpAddress()
		if err != nil {
			return nil, fmt.Errorf("error allocating IP address from subnet: %w", err)
		}

		fips = append(fips, ports.IP{
			SubnetID:  subnetInfo.subnet.ID,
			IPAddress: addr,
		})
	}

	up := true
	if opts.AdminStateUp != nil {
		up = *opts.AdminStateUp
	}

	port := &ports.Port{
		ID:           uuid.NewString(),
		Name:         opts.Name,
		NetworkID:    opts.NetworkID,
		DeviceID:     opts.DeviceID,
		FixedIPs:     fips,
		AdminStateUp: up,
		Status:       ports.StatusActive,
	}
	s.ports = append(s.ports, port)
	return port, nil
}

func (s *mainStore) DeletePort(ctx context.Context, id string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)
	return s.deletePortNoLock(ctx, id)
}

func (s *mainStore) deletePortNoLock(ctx context.Context, id string) error {
	var port *ports.Port
	for _, p := range s.ports {
		if p.ID == id {
			port = p
			break
		}
	}
	if port == nil {
		return sapmeta.NewNotFoundError("port not found")
	}

	for _, ip := range port.FixedIPs {
		if ip.SubnetID == "" {
			continue
		}
		subnetInfo := s.getSubnetInfoByIdNoLock(port.FixedIPs[0].SubnetID)
		if subnetInfo.subnet == nil {
			return sapmeta.NewNotFoundError("subnet not found for port")
		}
		err := subnetInfo.addressSpace.ReleaseIpAddress(ip.IPAddress)
		if err != nil {
			return fmt.Errorf("error releasing IP address %q back to subnet: %w", ip.IPAddress, err)
		}
	}

	s.ports = pie.FilterNot(s.ports, func(p *ports.Port) bool {
		return p.ID == id
	})
	return nil
}

func (s *mainStore) ListRouterSubnetInterfaces(ctx context.Context, routerId string) ([]sapclient.RouterSubnetInterfaceInfo, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	arrPorts, err := s.listPortsNoLock(ctx, ports.ListOpts{
		DeviceID: routerId,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing router ports: %w", err)
	}

	var result []sapclient.RouterSubnetInterfaceInfo
	for _, port := range arrPorts {
		if port.DeviceOwner != "network:router_gateway" {
			for _, ipSpec := range port.FixedIPs {
				result = append(result, sapclient.RouterSubnetInterfaceInfo{
					PortID:    port.ID,
					IpAddress: ipSpec.IPAddress,
					SubnetID:  ipSpec.SubnetID,
				})
			}
		}
	}
	return result, nil
}

// ShareClient implementation -------------------------------------------------

func (s *mainStore) ListShareNetworks(ctx context.Context, opts sharenetworks.ListOpts) ([]sharenetworks.ShareNetwork, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	var result []sharenetworks.ShareNetwork
	for _, arr := range s.shareNetworks {
		for _, sn := range arr {
			isMatch := true
			if opts.Name != "" && sn.Name != opts.Name {
				isMatch = false
			}
			if opts.NeutronNetID != "" && sn.NeutronNetID != opts.NeutronNetID {
				isMatch = false
			}
			if opts.NeutronSubnetID != "" && sn.NeutronSubnetID != opts.NeutronSubnetID {
				isMatch = false
			}
			if opts.IPVersion != 0 && sn.IPVersion != int(opts.IPVersion) {
				isMatch = false
			}
			if isMatch {
				result = append(result, *sn)
			}
		}
	}
	return result, nil
}

func (s *mainStore) ListShareNetworksByNetworkId(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	return s.ListShareNetworks(ctx, sharenetworks.ListOpts{
		NeutronNetID: networkId,
	})
}

func (s *mainStore) GetShareNetwork(ctx context.Context, shareNetworkId string) (*sharenetworks.ShareNetwork, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	for _, arr := range s.shareNetworks {
		for _, sn := range arr {
			if sn.ID == shareNetworkId {
				return sn, nil
			}
		}
	}
	return nil, nil
}

func (s *mainStore) CreateShareNetwork(ctx context.Context, opts sharenetworks.CreateOpts) (*sharenetworks.ShareNetwork, error) {
	if opts.NeutronNetID == "" {
		return nil, sapmeta.NewBadRequestError("neutron network id required")
	}
	if opts.NeutronSubnetID == "" {
		return nil, sapmeta.NewBadRequestError("neutron subnet id required")
	}
	if opts.Name == "" {
		return nil, sapmeta.NewBadRequestError("share network name required")
	}

	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	sn := &sharenetworks.ShareNetwork{
		ID:              uuid.NewString(),
		NeutronNetID:    opts.NeutronNetID,
		NeutronSubnetID: opts.NeutronSubnetID,
		Description:     opts.Description,
		Name:            opts.Name,
	}
	s.shareNetworks[opts.NeutronNetID] = append(s.shareNetworks[opts.NeutronNetID], sn)
	return sn, nil
}

func (s *mainStore) DeleteShareNetwork(ctx context.Context, shareNetworkId string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	if s.shares != nil {
		for _, arr := range s.shares {
			for _, sn := range arr {
				if sn.ShareNetworkID == shareNetworkId {
					return sapmeta.NewBadRequestError("can not delete share network with existing shares")
				}
			}
		}
	}

	for netId, arr := range s.shareNetworks {
		s.shareNetworks[netId] = pie.Filter(arr, func(x *sharenetworks.ShareNetwork) bool {
			return x.ID != shareNetworkId
		})
	}

	return nil
}

func (s *mainStore) ListShares(ctx context.Context, opts shares.ListOpts) ([]shares.Share, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	var result []shares.Share
	for _, arr := range s.shares {
		for _, share := range arr {
			isMatch := true
			if opts.Name != "" && share.Name != opts.Name {
				isMatch = false
			}
			if opts.Status != "" && share.Status != opts.Status {
				isMatch = false
			}
			if opts.ShareNetworkID != "" && share.ShareNetworkID != opts.ShareNetworkID {
				isMatch = false
			}
			if isMatch {
				result = append(result, *share)
			}
		}
	}
	return result, nil
}

func (s *mainStore) GetShare(ctx context.Context, id string) (*shares.Share, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	for _, arr := range s.shares {
		for _, share := range arr {
			if share.ID == id {
				return share, nil
			}
		}
	}
	return nil, nil
}

func (s *mainStore) CreateShare(ctx context.Context, opts shares.CreateOpts) (*shares.Share, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	id := uuid.NewString()
	sh := &shares.Share{
		ID:             id,
		ShareNetworkID: opts.ShareNetworkID,
		Name:           opts.Name,
		Size:           opts.Size,
		SnapshotID:     opts.SnapshotID,
		Metadata:       opts.Metadata,
		Status:         "creating",
	}
	s.shares[opts.ShareNetworkID] = append(s.shares[opts.ShareNetworkID], sh)
	return sh, nil
}

func (s *mainStore) DeleteShare(ctx context.Context, shareId string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	for netId, arr := range s.shares {
		s.shares[netId] = pie.Filter(arr, func(x *shares.Share) bool {
			return x.ID != shareId
		})
	}
	return nil
}

func (s *mainStore) ShareShrink(ctx context.Context, shareId string, newSize int) error {
	return s.shareChangeSize(ctx, shareId, newSize)
}

func (s *mainStore) ShareExtend(ctx context.Context, shareId string, newSize int) error {
	return s.shareChangeSize(ctx, shareId, newSize)
}

func (s *mainStore) shareChangeSize(ctx context.Context, shareId string, newSize int) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	var theShare *shares.Share
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

func (s *mainStore) ListShareExportLocations(ctx context.Context, id string) ([]shares.ExportLocation, error) {
	if util.IsContextDone(ctx) {
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

func (s *mainStore) ListShareAccessRules(ctx context.Context, id string) ([]sapclient.ShareAccess, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

	arr, ok := s.access[id]
	if !ok {
		return nil, nil
	}
	return deptrSlice(arr), nil
}

func (s *mainStore) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*sapclient.ShareAccess, error) {
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

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
	a := &sapclient.ShareAccess{
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

func (s *mainStore) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	if util.IsContextDone(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	time.Sleep(time.Millisecond)

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
	arr = pie.Filter(arr, func(x *sapclient.ShareAccess) bool {
		return x.ID != accessId
	})
	s.access[shareId] = arr
	return nil
}

func (s *mainStore) CreateShareNetworkOp(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	return s.CreateShareNetwork(ctx, sharenetworks.CreateOpts{
		NeutronNetID:    networkId,
		NeutronSubnetID: subnetId,
		Name:            name,
	})
}

func (s *mainStore) ListSharesInShareNetwork(ctx context.Context, shareNetworkId string) ([]shares.Share, error) {
	return s.ListShares(ctx, shares.ListOpts{ShareNetworkID: shareNetworkId})
}

func (s *mainStore) CreateShareOp(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error) {
	return s.CreateShare(ctx, shares.CreateOpts{
		ShareNetworkID: shareNetworkId,
		Name:           name,
		Size:           size,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
	})
}
