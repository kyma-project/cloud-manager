package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

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

// WebAclClient adapter methods - wrap webAclStore methods to match Client interface

func (s *accountRegionStore) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*wafv2types.WebACL, string, error) {
	return s.webAclStore.CreateWebACL(ctx, input)
}

func (s *accountRegionStore) UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error {
	return s.webAclStore.UpdateWebACL(ctx, input)
}

func (s *accountRegionStore) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	return s.webAclStore.DeleteWebACL(ctx, name, id, scope, lockToken)
}

func (s *accountRegionStore) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	return s.webAclStore.GetWebACL(ctx, name, id, scope)
}

func (s *accountRegionStore) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	return s.webAclStore.ListWebACLs(ctx, scope)
}
