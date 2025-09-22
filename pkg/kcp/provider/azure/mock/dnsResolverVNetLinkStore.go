package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ DnsResolverVNetLinkClient = &dnsResolverVNetLinkStore{}

type dnsResolverVNetLinkStore struct {
	m            sync.Mutex
	subscription string

	// items are resourceGroupName => dnsForwardingRulesetName => virtualNetworkLinkName
	items map[string]map[string]map[string]*armdnsresolver.VirtualNetworkLink
}

func newDnsResolverVNetLinkStore(subscription string) *dnsResolverVNetLinkStore {
	return &dnsResolverVNetLinkStore{
		subscription: subscription,
		items:        make(map[string]map[string]map[string]*armdnsresolver.VirtualNetworkLink),
	}
}

func (s *dnsResolverVNetLinkStore) CreateDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, vnetId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.items[resourceGroupName]; !ok {
		s.items[resourceGroupName] = map[string]map[string]*armdnsresolver.VirtualNetworkLink{}
	}

	if _, ok := s.items[resourceGroupName][dnsForwardingRulesetName]; !ok {
		s.items[resourceGroupName][dnsForwardingRulesetName] = map[string]*armdnsresolver.VirtualNetworkLink{}
	}

	item := &armdnsresolver.VirtualNetworkLink{
		ID: ptr.To(azureutil.NewDnsResolverVirtualNetworkLinkResourceId(s.subscription, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName).String()),
		Properties: &armdnsresolver.VirtualNetworkLinkProperties{
			VirtualNetwork: &armdnsresolver.SubResource{ID: ptr.To(vnetId)},
		},
		Name: ptr.To(virtualNetworkLinkName),
	}

	s.items[resourceGroupName][dnsForwardingRulesetName][virtualNetworkLinkName] = item
	return nil
}
func (s *dnsResolverVNetLinkStore) GetDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	link, err := s.getDnsResolverVNetLinkNonLocking(resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName)

	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(link)

	if err != nil {
		return nil, err
	}

	return res, nil
}
func (s *dnsResolverVNetLinkStore) DeleteDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) error {

	if isContextCanceled(ctx) {
		return context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	delete(s.items[resourceGroupName][dnsForwardingRulesetName], virtualNetworkLinkName)

	return nil
}

func (s *dnsResolverVNetLinkStore) getDnsResolverVNetLinkNonLocking(resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error) {
	if group, ok := s.items[resourceGroupName]; ok {
		if ruleset, ok := group[dnsForwardingRulesetName]; ok {
			if link, ok := ruleset[virtualNetworkLinkName]; ok {
				result, err := util.JsonClone(link)
				return result, err
			}
		}
	}

	return nil, azuremeta.NewAzureNotFoundError()
}
