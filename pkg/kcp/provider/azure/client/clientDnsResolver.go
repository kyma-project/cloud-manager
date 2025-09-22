package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
)

type DnsForwardingRulesetsClient interface {
	Get(ctx context.Context, resourceGroupName string, dnsForwardingRulesetName string) (*armdnsresolver.DNSForwardingRuleset, error)
}

func NewDnsForwardingRulesetsClient(svc *armdnsresolver.DNSForwardingRulesetsClient) DnsForwardingRulesetsClient {
	return &dnsForwardingRulesetsClient{svc: svc}
}

type dnsForwardingRulesetsClient struct {
	svc *armdnsresolver.DNSForwardingRulesetsClient
}

func (c *dnsForwardingRulesetsClient) Get(ctx context.Context, resourceGroupName string, dnsForwardingRulesetName string) (*armdnsresolver.DNSForwardingRuleset, error) {
	response, err := c.svc.Get(ctx, resourceGroupName, dnsForwardingRulesetName, nil)
	if err != nil {
		return nil, err
	}
	return &response.DNSForwardingRuleset, nil
}
