package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
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
	*certificateStore

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
		certificateStore:      newCertificateStore(account, region),
	}
}

func (s *accountRegionStore) Region() string {
	return s.region
}

// WebAclClient adapter methods - wrap webAclStore methods to match Client interface

func (s *accountRegionStore) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) error {
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

// AcmClient adapter methods - wrap certificateStore methods to match Client interface

func (s *accountRegionStore) ImportCertificate(ctx context.Context, input *acm.ImportCertificateInput) (string, error) {
	return s.certificateStore.ImportCertificate(ctx, input)
}

func (s *accountRegionStore) DescribeCertificate(ctx context.Context, arn string) (*acmtypes.CertificateDetail, error) {
	return s.certificateStore.DescribeCertificate(ctx, arn)
}

func (s *accountRegionStore) DeleteCertificate(ctx context.Context, arn string) error {
	return s.certificateStore.DeleteCertificate(ctx, arn)
}

func (s *accountRegionStore) GetCertificate(ctx context.Context, arn string) (string, string, error) {
	return s.certificateStore.GetCertificate(ctx, arn)
}

func (s *accountRegionStore) SearchCertificates(ctx context.Context, input *acm.SearchCertificatesInput) ([]acmtypes.CertificateSearchResult, error) {
	return s.certificateStore.SearchCertificates(ctx, input)
}

func (s *accountRegionStore) ListTagsForCertificate(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	return s.certificateStore.ListTagsForCertificate(ctx, arn)
}
