package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/3th1nk/cidr"
	"github.com/kyma-project/cloud-manager/pkg/common"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

type ExposedDataConfig interface {
	InsertVpcRouter(project string, region string, router *computepb.Router) (*computepb.Router, error)
	InsertAddress(project string, region string, address *computepb.Address) (*computepb.Address, error)
}

type exposedDataStore struct {
	m sync.Mutex

	ipPool iprangeallocate.AddressSpace

	// project => region => [router]
	routers map[string]map[string][]*computepb.Router
	// project => region => [address]
	addresses map[string]map[string][]*computepb.Address
}

var _ gcpexposeddataclient.Client = (*exposedDataStore)(nil)
var _ ExposedDataConfig = (*exposedDataStore)(nil)

// ExposedDataConfig =============================================

func (s *exposedDataStore) InsertVpcRouter(project string, region string, router *computepb.Router) (*computepb.Router, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if router == nil {
		return nil, fmt.Errorf("router nil: %w", common.ErrLogical)
	}
	if router.Name == nil || ptr.Deref(router.Name, "") == "" {
		return nil, fmt.Errorf("router name empty: %w", common.ErrLogical)
	}
	if router.Network == nil || ptr.Deref(router.Network, "") == "" {
		return nil, fmt.Errorf("router network empty: %w", common.ErrLogical)
	}
	if !strings.Contains(ptr.Deref(router.Network, ""), "/compute/v1/projects/") {
		router.Network = ptr.To(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, ptr.Deref(router.Network, "")))
	}

	if router.CreationTimestamp == nil {
		router.CreationTimestamp = ptr.To(time.Now().Format(time.RFC3339))
	}
	router.Region = ptr.To(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s", project, region))
	router.SelfLink = ptr.To(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/routers/%s", project, region, ptr.Deref(router.Name, "")))

	// -----------

	if s.routers == nil {
		s.routers = map[string]map[string][]*computepb.Router{}
	}
	if _, ok := s.routers[project]; !ok {
		s.routers[project] = map[string][]*computepb.Router{}
	}

	s.routers[project][region] = append(s.routers[project][region], router)

	return router, nil
}

func (s *exposedDataStore) InsertAddress(project string, region string, address *computepb.Address) (*computepb.Address, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if address == nil {
		return nil, fmt.Errorf("address nil: %w", common.ErrLogical)
	}
	if address.Name == nil || ptr.Deref(address.Name, "") == "" {
		return nil, fmt.Errorf("address name empty: %w", common.ErrLogical)
	}
	if address.Purpose == nil || ptr.Deref(address.Purpose, "") == "" {
		return nil, fmt.Errorf("address purpose empty: %w", common.ErrLogical)
	}

	if address.CreationTimestamp == nil {
		address.CreationTimestamp = ptr.To(time.Now().Format(time.RFC3339))
	}
	if address.Address == nil {
		ip, err := s.ipPool.AllocateWithPreference(32, "33.0.0.0/32")
		if err != nil {
			return nil, fmt.Errorf("ip allocation failed: %w", err)
		}
		address.Address = ptr.To(cidr.ParseNoError(ip).CIDR().IP.String())
	}
	address.Status = ptr.To("IN_USE")
	address.Region = ptr.To(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s", project, region))
	address.SelfLink = ptr.To(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/addresses/%s", project, region, ptr.Deref(address.Name, "")))

	// -----------

	if s.addresses == nil {
		s.addresses = map[string]map[string][]*computepb.Address{}
	}
	if _, ok := s.addresses[project]; !ok {
		s.addresses[project] = map[string][]*computepb.Address{}
	}

	s.addresses[project][region] = append(s.addresses[project][region], address)

	return address, nil
}

// gcpexposeddataclient.Client ===================================

func (s *exposedDataStore) GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}

	if _, ok := s.routers[project]; !ok {
		return nil, nil
	}
	if _, ok := s.routers[project][region]; !ok {
		return nil, nil
	}

	var results []*computepb.Router
	for _, router := range s.routers[project][region] {
		if strings.Contains(ptr.Deref(router.Network, ""), vpcName) {
			results = append(results, util.Must(util.JsonClone(router)))
		}
	}

	return results, nil
}

func (s *exposedDataStore) GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, context.Canceled
	}

	if _, ok := s.addresses[project]; !ok {
		return nil, nil
	}
	if _, ok := s.addresses[project][region]; !ok {
		return nil, nil
	}

	var results []*computepb.Address
	for _, address := range s.addresses[project][region] {
		for _, usr := range address.Users {
			if strings.Contains(usr, routerName) {
				results = append(results, util.Must(util.JsonClone(address)))
			}
		}
	}

	return results, nil
}
