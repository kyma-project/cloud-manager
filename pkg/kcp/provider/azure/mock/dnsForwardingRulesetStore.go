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

var _ DnsForwardingRulesetClient = &dnsForwardingRulesetStore{}
var _ DnsForwardingRulesetConfig = &dnsForwardingRulesetStore{}

type dnsForwardingRulesetStore struct {
	m            sync.Mutex
	subscription string

	// items are resourceGroupName => dnsForwardingRulesetName
	items map[string]map[string]*armdnsresolver.DNSForwardingRuleset
}

func newDnsForwardingRulesetStore(subscription string) *dnsForwardingRulesetStore {
	return &dnsForwardingRulesetStore{
		subscription: subscription,
		items:        make(map[string]map[string]*armdnsresolver.DNSForwardingRuleset),
	}
}

func (s *dnsForwardingRulesetStore) Get(ctx context.Context, resourceGroupName string, dnsForwardingRulesetName string) (*armdnsresolver.DNSForwardingRuleset, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	if ruleset, ok := s.items[resourceGroupName][dnsForwardingRulesetName]; ok {
		return util.JsonClone(ruleset)
	}

	return nil, azuremeta.NewAzureNotFoundError()
}

func (s *dnsForwardingRulesetStore) CreateDnsForwardingRuleset(ctx context.Context, resourceGroup, dnsForwardingRulesetName string, tags map[string]string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.items[resourceGroup]; !ok {
		s.items[resourceGroup] = make(map[string]*armdnsresolver.DNSForwardingRuleset)
	}

	item := &armdnsresolver.DNSForwardingRuleset{
		ID:   ptr.To(azureutil.NewDnsForwardingRulesetResourceId(s.subscription, resourceGroup, dnsForwardingRulesetName).String()),
		Name: ptr.To(dnsForwardingRulesetName),
		Tags: azureutil.AzureTags(tags),
	}

	s.items[resourceGroup][dnsForwardingRulesetName] = item
	return nil
}
