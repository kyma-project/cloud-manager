package mock

var _ Clients = &tenantSubscriptionStore{}
var _ Configs = &tenantSubscriptionStore{}
var _ TenantSubscription = &tenantSubscriptionStore{}

type tenantSubscriptionStore struct {
	*resourceStore
	*networkStore
	*securityGroupsStore
	*redisStore
	*privateEndPointsStore
	*privateDnsZoneStore
	*virtualNetworkLinkStore
	*privateDnsZoneGroupStore
	*storageStore
	*natGatewayStore
	*publicIpAddressStore
	*fileShareStore
	*dnsResolverVNetLinkStore
	*dnsForwardingRulesetStore
	*networkFlowLogsStore
	*operationalInsightsStore
	*securityStore
	*storageAccountStore

	tenant       string
	subscription string

	server Server
}

func newTenantSubscriptionStore(tenant, subscription string, srv Server) *tenantSubscriptionStore {
	return &tenantSubscriptionStore{
		server:                    srv,
		resourceStore:             newResourceStore(subscription),
		networkStore:              newNetworkStore(subscription),
		securityGroupsStore:       newSecurityGroupsStore(subscription),
		redisStore:                newRedisStore(subscription),
		privateEndPointsStore:     newPrivateEndPointsStore(subscription),
		privateDnsZoneStore:       newPrivateDnsZoneStore(subscription),
		virtualNetworkLinkStore:   newVirtualNetworkLinkStore(subscription),
		privateDnsZoneGroupStore:  newPrivateDnsZoneGroupStore(subscription),
		storageStore:              newStorageStore(subscription),
		natGatewayStore:           newNatGatewayStore(subscription),
		publicIpAddressStore:      newPublicIpAddressStore(subscription),
		fileShareStore:            newFileShareStore(subscription),
		dnsForwardingRulesetStore: newDnsForwardingRulesetStore(subscription),
		dnsResolverVNetLinkStore:  newDnsResolverVNetLinkStore(subscription),
		networkFlowLogsStore:      newNetworkFlowLogsStore(subscription),
		operationalInsightsStore:  newOperationalInsightsStore(subscription),
		securityStore:             newSecurityStore(subscription),
		storageAccountStore:       newStorageAccountStore(subscription),
		tenant:                    tenant,
		subscription:              subscription,
	}
}

func (s *tenantSubscriptionStore) TenantId() string {
	return s.tenant
}

func (s *tenantSubscriptionStore) SubscriptionId() string {
	return s.subscription
}

func (s *tenantSubscriptionStore) Equals(other TenantSubscription) bool {
	return s.SubscriptionId() == other.SubscriptionId() && s.TenantId() == other.TenantId()
}

func (s *tenantSubscriptionStore) Delete() {
	s.server.DeleteSubscription(s)
}
