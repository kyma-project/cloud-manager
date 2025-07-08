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
	tenant       string
	subscription string
}

func newTenantSubscriptionStore(tenant, subscription string) *tenantSubscriptionStore {
	return &tenantSubscriptionStore{
		resourceStore:            newResourceStore(subscription),
		networkStore:             newNetworkStore(subscription),
		securityGroupsStore:      newSecurityGroupsStore(subscription),
		redisStore:               newRedisStore(subscription),
		privateEndPointsStore:    newPrivateEndPointsStore(subscription),
		privateDnsZoneStore:      newPrivateDnsZoneStore(subscription),
		virtualNetworkLinkStore:  newVirtualNetworkLinkStore(subscription),
		privateDnsZoneGroupStore: newPrivateDnsZoneGroupStore(subscription),
		storageStore:             newStorageStore(subscription),
		natGatewayStore:          newNatGatewayStore(subscription),
		publicIpAddressStore:     newPublicIpAddressStore(subscription),
		fileShareStore:           newFileShareStore(subscription),
		tenant:                   tenant,
		subscription:             subscription,
	}
}

func (s *tenantSubscriptionStore) TenantId() string {
	return s.tenant
}

func (s *tenantSubscriptionStore) SubscriptionId() string {
	return s.subscription
}
