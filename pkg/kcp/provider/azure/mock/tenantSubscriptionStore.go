package mock

var _ Clients = &tenantSubscriptionStore{}
var _ Configs = &tenantSubscriptionStore{}
var _ TenantSubscription = &tenantSubscriptionStore{}

type tenantSubscriptionStore struct {
	*networkStore
	*redisStore
	tenant       string
	subscription string
}

func newTenantSubscriptionStore(tenant, subscription string) *tenantSubscriptionStore {
	return &tenantSubscriptionStore{
		tenant:       tenant,
		subscription: subscription,
		networkStore: newNetworkStore(subscription),
		redisStore:   newRedisStore(subscription),
	}
}
