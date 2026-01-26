package mock

var _ AccountRegion = (*accountRegionStore)(nil)

type accountRegionStore struct {
	*vpcStore
	*nfsStore
	*vpcPeeringStore
	*elastiCacheClientFake
	*routeTablesStore

	region string
}

func newAccountRegionStore(region string) *accountRegionStore {
	return &accountRegionStore{
		region: region,

		vpcStore:              newVpcStore(),
		vpcPeeringStore:       newVpcPeeringStore(),
		elastiCacheClientFake: newElastiCacheClientFake(),
		nfsStore:              &nfsStore{},
		routeTablesStore:      &routeTablesStore{},
	}
}

func (s *accountRegionStore) Region() string {
	return s.region
}
