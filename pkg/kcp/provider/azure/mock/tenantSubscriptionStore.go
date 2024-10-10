package mock

var _ Clients = &tenantSubscriptionStore{}
var _ Configs = &tenantSubscriptionStore{}
var _ TenantSubscription = &tenantSubscriptionStore{}

type tenantSubscriptionStore struct {
	*resourceStore
	*networkStore
	*securityGroupsStore
	*redisStore
	*privateDnsZoneStore
	*virtualNetworkLinkStore
	tenant       string
	subscription string
}

func newTenantSubscriptionStore(tenant, subscription string) *tenantSubscriptionStore {
	return &tenantSubscriptionStore{
		resourceStore:           newResourceStore(subscription),
		networkStore:            newNetworkStore(subscription),
		securityGroupsStore:     newSecurityGroupsStore(subscription),
		redisStore:              newRedisStore(subscription),
		privateDnsZoneStore:     newPrivateDnsZoneStore(subscription),
		virtualNetworkLinkStore: newVirtualNetworkLinkStore(subscription),
		tenant:                  tenant,
		subscription:            subscription,
	}
}
