package mock

import (
	"context"

	logstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
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
	*logsStore

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
		logsStore:             newLogsStore(),
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

// CloudWatch Logs adapter methods
func (s *accountRegionStore) CreateLogGroup(ctx context.Context, logGroupName string) error {
	return s.logsStore.CreateLogGroup(ctx, logGroupName)
}

func (s *accountRegionStore) DeleteLogGroup(ctx context.Context, logGroupName string) error {
	return s.logsStore.DeleteLogGroup(ctx, logGroupName)
}

func (s *accountRegionStore) GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error) {
	return s.logsStore.GetLogGroup(ctx, logGroupName)
}

func (s *accountRegionStore) DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error) {
	return s.logsStore.DescribeLogGroups(ctx, prefix)
}

func (s *accountRegionStore) PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error {
	return s.logsStore.PutRetentionPolicy(ctx, logGroupName, retentionDays)
}

func (s *accountRegionStore) TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error {
	return s.logsStore.TagLogGroup(ctx, logGroupName, tags)
}
