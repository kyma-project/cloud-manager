package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
	"k8s.io/utils/ptr"
	"time"
)

type DnsResolverVNetLinkClient interface {
	CreateDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, vnetId string) error
	GetDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error)
	DeleteDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) error
}

func NewDnsResolverVNetLinkClient(svc *armdnsresolver.VirtualNetworkLinksClient) DnsResolverVNetLinkClient {
	return &dnsResolverVNetLinkClient{svc: svc}
}

type dnsResolverVNetLinkClient struct {
	svc *armdnsresolver.VirtualNetworkLinksClient
}

func (c *dnsResolverVNetLinkClient) CreateDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, vnetId string) error {

	parameters := armdnsresolver.VirtualNetworkLink{
		Properties: &armdnsresolver.VirtualNetworkLinkProperties{
			VirtualNetwork: &armdnsresolver.SubResource{
				ID: ptr.To(vnetId),
			},
		},
	}

	poller, err := c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, parameters, nil)

	if err != nil {
		return err
	}

	// operation usually takes around 5 seconds but a minute is given
	pollerCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	_, err = poller.PollUntilDone(pollerCtx, nil)

	return err
}

func (c *dnsResolverVNetLinkClient) GetDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) (*armdnsresolver.VirtualNetworkLink, error) {

	r, err := c.svc.Get(ctx, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, nil)

	if err != nil {
		return nil, err
	}

	return &r.VirtualNetworkLink, nil
}

func (c *dnsResolverVNetLinkClient) DeleteDnsResolverVNetLink(ctx context.Context, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, dnsForwardingRulesetName, virtualNetworkLinkName, nil)
	return err
}
