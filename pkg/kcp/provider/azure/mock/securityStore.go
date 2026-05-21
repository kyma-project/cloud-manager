package mock

import (
	"context"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/elliotchance/pie/v2"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func newSecurityStore(subscription string) *securityStore {
	return &securityStore{
		subscription: subscription,
	}
}

type securityStore struct {
	m            sync.Mutex
	subscription string

	pricings []*armsecurity.Pricing
}

var _ azureclient.SecurityClient = (*securityStore)(nil)

func (s *securityStore) ListSecurityPricings(ctx context.Context, scopeID string, options *armsecurity.PricingsClientListOptions) (armsecurity.PricingsClientListResponse, error) {
	var result armsecurity.PricingsClientListResponse
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, pricing := range s.pricings {
		result.Value = append(result.Value, util.Must(util.Clone(pricing)))
	}
	return result, nil
}

func (s *securityStore) UpdateSecurityPricing(ctx context.Context, scopeID string, pricingName string, pricing armsecurity.Pricing, options *armsecurity.PricingsClientUpdateOptions) (armsecurity.PricingsClientUpdateResponse, error) {
	var result armsecurity.PricingsClientUpdateResponse
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	s.pricings = pie.FilterNot(s.pricings, func(pricing *armsecurity.Pricing) bool {
		return ptr.Deref(pricing.Name, "") == pricingName
	})

	cpy := util.Must(util.Clone(&pricing))
	cpy.Name = new(pricingName)
	s.pricings = append(s.pricings, cpy)

	result.Pricing = *cpy

	return result, nil
}
