package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	CreateDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, vnetId string) error
	GetDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error)
	DeleteDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) error
	Get(ctx context.Context, resourceGroupName string, dnsForwardingRulesetName string) (*armdnsresolver.DNSForwardingRuleset, error)
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {

		credentialOptions := azureclient.NewCredentialOptions()
		credentialOptions.AdditionallyAllowedTenants = []string{"*"}

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, credentialOptions)

		if err != nil {
			return nil, err
		}

		options := azureclient.NewClientOptions()
		options.AuxiliaryTenants = auxiliaryTenants

		dnsResolverClientFactory, err := armdnsresolver.NewClientFactory(subscriptionId, cred, options)

		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewDnsResolverVNetLinkClient(
				dnsResolverClientFactory.NewVirtualNetworkLinksClient(),
			),

			azureclient.NewDnsForwardingRulesetsClient(
				dnsResolverClientFactory.NewDNSForwardingRulesetsClient(),
			),
		), nil
	}
}

type client struct {
	azureclient.DnsResolverVNetLinkClient
	azureclient.DnsForwardingRulesetsClient
}

func newClient(
	dnsResolverVNetLinkClient azureclient.DnsResolverVNetLinkClient,
	dnsForwardingRulesetsClient azureclient.DnsForwardingRulesetsClient) Client {
	return &client{
		DnsResolverVNetLinkClient:   dnsResolverVNetLinkClient,
		DnsForwardingRulesetsClient: dnsForwardingRulesetsClient,
	}
}
