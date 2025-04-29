package mock

type accountRegionStore struct {
	*vpcStore
	*nfsStore
	*vpcPeeringStore
	*elastiCacheClientFake
	*routeTablesStore
	*scopeStore
}

func newAccountRegionStore(account, region string) *accountRegionStore {
	return &accountRegionStore{
		scopeStore:            &scopeStore{account: account},
		vpcStore:              newVpcStore(account, region),
		vpcPeeringStore:       newVpcPeeringStore(),
		elastiCacheClientFake: newElastiCacheClientFake(),
		nfsStore:              &nfsStore{},
		routeTablesStore:      &routeTablesStore{},
	}
}
