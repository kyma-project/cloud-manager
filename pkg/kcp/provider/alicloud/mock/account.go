package mock

import (
	"sync"
)

var _ Account = (*account)(nil)

func newAccount(s *server, accountId, accessKeyId, accessKeySecret string) *account {
	return &account{
		server:    s,
		accountId: accountId,
		credentials: AccountCredential{
			AccessKeyId:     accessKeyId,
			AccessKeySecret: accessKeySecret,
		},
		regionalStores: map[string]*accountRegionStore{},
	}
}

type account struct {
	m sync.Mutex

	server      *server
	accountId   string
	credentials AccountCredential

	regionalStores map[string]*accountRegionStore
}

func (a *account) AccountId() string              { return a.accountId }
func (a *account) Credentials() AccountCredential { return a.credentials }

func (a *account) Region(region string) AccountRegion {
	a.m.Lock()
	defer a.m.Unlock()
	reg, ok := a.regionalStores[region]
	if !ok {
		reg = newAccountRegionStore(region)
		a.regionalStores[region] = reg
	}
	return reg
}

func (a *account) Delete() {
	a.server.deleteAccount(a.accountId)
}
