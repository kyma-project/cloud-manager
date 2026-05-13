package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

type SecurityClient interface {
	ListSecurityPricings(ctx context.Context, scopeID string, options *armsecurity.PricingsClientListOptions) (armsecurity.PricingsClientListResponse, error)
	UpdateSecurityPricing(ctx context.Context, scopeID string, pricingName string, pricing armsecurity.Pricing, options *armsecurity.PricingsClientUpdateOptions) (armsecurity.PricingsClientUpdateResponse, error)
}

func NewSecurityClient(svcPricings *armsecurity.PricingsClient) SecurityClient {
	return &securityClient{svcPricings: svcPricings}
}

var _ SecurityClient = (*securityClient)(nil)

type securityClient struct {
	svcPricings *armsecurity.PricingsClient
}

func (c *securityClient) ListSecurityPricings(ctx context.Context, scopeID string, options *armsecurity.PricingsClientListOptions) (armsecurity.PricingsClientListResponse, error) {
	return c.svcPricings.List(ctx, scopeID, options)
}

func (c *securityClient) UpdateSecurityPricing(ctx context.Context, scopeID string, pricingName string, pricing armsecurity.Pricing, options *armsecurity.PricingsClientUpdateOptions) (armsecurity.PricingsClientUpdateResponse, error) {
	return c.svcPricings.Update(ctx, scopeID, pricingName, pricing, options)
}
