package mock

var _ Clients = &tenantSubscriptionStore{}
var _ Configs = &tenantSubscriptionStore{}
var _ TenantSubscription = &tenantSubscriptionStore{}

type tenantSubscriptionStore struct {
	*resourceStore
	*networkStore
	*redisStore
	tenant       string
	subscription string
}

func newTenantSubscriptionStore(tenant, subscription string) *tenantSubscriptionStore {
	return &tenantSubscriptionStore{
		resourceStore: newResourceStore(subscription),
		networkStore:  newNetworkStore(subscription),
		redisStore:    newRedisStore(subscription),
		tenant:        tenant,
		subscription:  subscription,
	}
}
