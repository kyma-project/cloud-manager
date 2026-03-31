package mock

var _ AccountRegion = (*accountRegionStore)(nil)

type accountRegionStore struct {
	*vpcStore
	*nfsStore
	*vpcPeeringStore
	*elastiCacheClientFake
	*routeTablesStore
	*webAclStore

	region  string
	account string
}

func newAccountRegionStore(account, region string) *accountRegionStore {
	return &accountRegionStore{
		region:                region,
		account:               account,
		vpcStore:              newVpcStore(),
		vpcPeeringStore:       newVpcPeeringStore(),
		elastiCacheClientFake: newElastiCacheClientFake(),
		nfsStore:              &nfsStore{},
		routeTablesStore:      &routeTablesStore{},
		webAclStore:           newWebAclStore(account, region),
	}
}

func (s *accountRegionStore) Region() string {
	return s.region
}
