package mock

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/utils/ptr"
)

type StsClient interface {
	GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}

type Account interface {
	StsClient
	ScopeClient
	SubscriptionClient

	AccountId() string
	Credentials() AccountCredential

	Delete()

	Region(region string) AccountRegion
}

func newAccount(s *server, accountId, accessKeyId, secretAccessKey string) *account {
	return &account{
		server:    s,
		accountId: accountId,
		credentials: []AccountCredential{
			{
				AccessKeyId:     accessKeyId,
				SecretAccessKey: secretAccessKey,
			},
		},
		regionalStores: map[string]*accountRegionStore{},
	}
}

type account struct {
	m sync.Mutex

	server *server

	accountId      string
	credentials    []AccountCredential
	regionalStores map[string]*accountRegionStore
}

type AccountCredential struct {
	AccessKeyId     string
	SecretAccessKey string
}

func (a *account) Delete() {
	a.server.deleteAccount(a.accountId)
}

func (a *account) AccountId() string {
	return a.accountId
}

func (a *account) Credentials() AccountCredential {
	return a.credentials[0]
}

func (a *account) Region(region string) AccountRegion {
	a.m.Lock()
	defer a.m.Unlock()

	reg, ok := a.regionalStores[region]
	if !ok {
		reg = newAccountRegionStore(a.accountId, region)
		a.regionalStores[region] = reg
	}

	return reg
}

func (a *account) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	return &sts.GetCallerIdentityOutput{
		Account: ptr.To(a.accountId),
		Arn:     nil,
		UserId:  nil,
	}, nil
}
