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

	link, err := s.getVirtualNetworkLinkNonLocking(resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName)

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
	return nil
}

func (s *dnsResolverVNetLinkStore) getVirtualNetworkLinkNonLocking(resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	ruleset, ok := group[dnsForwardingRulesetName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	virtualNetworkLink, ok := ruleset[virtualNetworkLinkName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	if virtualNetworkLink == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return virtualNetworkLink, nil
}
