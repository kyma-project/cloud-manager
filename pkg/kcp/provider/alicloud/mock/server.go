package mock

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	alicloudredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	alicloudredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

var _ Server = (*server)(nil)

var ErrInvalidCredentials = errors.New("alicloud mock: invalid credentials")

// New returns a fresh in-memory Alicloud mock server.
func New() Server {
	return &server{accounts: map[string]*account{}}
}

type server struct {
	m        sync.Mutex
	accounts map[string]*account
}

func (s *server) NewAccount() Account {
	return s.NewAccountWithCredentials("ak-"+uuid.NewString()[:8], "sk-"+uuid.NewString())
}

func (s *server) NewAccountWithCredentials(accessKeyId, accessKeySecret string) Account {
	s.m.Lock()
	defer s.m.Unlock()
	id := uuid.NewString()
	a := newAccount(s, id, accessKeyId, accessKeySecret)
	s.accounts[id] = a
	return a
}

func (s *server) GetAccount(accountId string) Account {
	s.m.Lock()
	defer s.m.Unlock()
	a, ok := s.accounts[accountId]
	if !ok {
		return nil
	}
	return a
}

func (s *server) Login(accessKeyId, accessKeySecret string) (Account, error) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, a := range s.accounts {
		if a.credentials.AccessKeyId == accessKeyId && a.credentials.AccessKeySecret == accessKeySecret {
			return a, nil
		}
	}
	return nil, ErrInvalidCredentials
}

func (s *server) deleteAccount(accountId string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.accounts, accountId)
}

func (s *server) IpRangeClientProvider() alicloudiprangeclient.ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (alicloudiprangeclient.Client, error) {
		// In tests the state factory passes credentials from AlicloudConfig which are empty.
		// Fall back to the first registered account so mock tests work without real credentials.
		a, err := s.Login(accessKeyId, accessKeySecret)
		if err != nil {
			a = s.firstAccount()
		}
		if a == nil {
			return nil, ErrInvalidCredentials
		}
		return a.Region(region).IpRangeClient(), nil
	}
}

func (s *server) VpcNetworkClientProvider() alicloudvpcnetworkclient.ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (alicloudvpcnetworkclient.Client, error) {
		a, err := s.Login(accessKeyId, accessKeySecret)
		if err != nil {
			a = s.firstAccount()
		}
		if a == nil {
			return nil, ErrInvalidCredentials
		}
		return a.Region(region).VpcNetworkClient(), nil
	}
}

func (s *server) NfsInstanceClientProvider() alicloudnfsinstanceclient.ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (alicloudnfsinstanceclient.Client, error) {
		a, err := s.Login(accessKeyId, accessKeySecret)
		if err != nil {
			a = s.firstAccount()
		}
		if a == nil {
			return nil, ErrInvalidCredentials
		}
		return a.Region(region).NfsInstanceClient(), nil
	}
}

func (s *server) RedisInstanceClientProvider() alicloudredisinstanceclient.ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (alicloudredisinstanceclient.Client, error) {
		a, err := s.Login(accessKeyId, accessKeySecret)
		if err != nil {
			a = s.firstAccount()
		}
		if a == nil {
			return nil, ErrInvalidCredentials
		}
		return a.Region(region).RedisInstanceClient(), nil
	}
}

func (s *server) RedisClusterClientProvider() alicloudredisclusterclient.ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (alicloudredisclusterclient.Client, error) {
		a, err := s.Login(accessKeyId, accessKeySecret)
		if err != nil {
			a = s.firstAccount()
		}
		if a == nil {
			return nil, ErrInvalidCredentials
		}
		return a.Region(region).RedisClusterClient(), nil
	}
}

func (s *server) firstAccount() Account {
	s.m.Lock()
	defer s.m.Unlock()
	for _, a := range s.accounts {
		return a
	}
	return nil
}
