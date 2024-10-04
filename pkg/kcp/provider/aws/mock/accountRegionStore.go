package mock

type accountRegionStore struct {
	*vpcStore
	*nfsStore
	*vpcPeeringStore
	*elastiCacheClientFake
	*routeTablesStore
}

func newAccountRegionStore(account, region string) *accountRegionStore {
	return &accountRegionStore{
		vpcStore:              newVpcStore(),
		vpcPeeringStore:       newVpcPeeringStore(),
		elastiCacheClientFake: newElastiCacheClientFake(),
		nfsStore:              &nfsStore{},
		routeTablesStore:      &routeTablesStore{},
	}
}
