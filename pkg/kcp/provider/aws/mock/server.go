package mock

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/aws/smithy-go"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"

	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

var _ Server = &server{}

var ErrNoAccount = errors.New("no aws account found")

func New() Server {
	return &server{
		accountStores: map[string]*accountRegionStore{},
	}
}

type server struct {
	m sync.Mutex

	accounts []*account

	accountStores map[string]*accountRegionStore
}

func (s *server) NewAccount() Account {
	s.m.Lock()
	defer s.m.Unlock()

	for i := 0; i < 10; i++ {
		accountId := fmt.Sprintf("%d", rand.Int31())
		taken := false
		for _, acc := range s.accounts {
			if acc.accountId == accountId {
				taken = true
				break
			}
		}
		if !taken {
			accessKeyId := uuid.NewString()
			secretAccessKey := uuid.NewString()
			acc := newAccount(s, accountId, accessKeyId, secretAccessKey)
			s.accounts = append(s.accounts, acc)
			return acc
		}
	}
	return nil
}

func (s *server) deleteAccount(accountId string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.accounts = pie.FilterNot(s.accounts, func(acc *account) bool {
		return acc.accountId == accountId
	})
}

func (a *server) GetAccount(accountId string) Account {
	for _, acc := range a.accounts {
		if acc.accountId == accountId {
			return acc
		}
	}
	return nil
}

func (a *server) Login(key, secret string) (Account, error) {
	if key == "" || secret == "" {
		return nil, fmt.Errorf("invalid empty aws credentials")
	}

	for _, acc := range a.accounts {
		for _, cred := range acc.credentials {
			if cred.AccessKeyId == key && cred.SecretAccessKey == secret {
				return acc, nil
			}
		}
	}

	return nil, &smithy.GenericAPIError{
		Code:    awsmeta.AccessDenied,
		Message: "invalid aws credentials",
	}
}

// Providers ==================================================================

func (s *server) ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (scopeclient.AwsStsClient, error) {
		acc, err := s.Login(key, secret)
		if err != nil {
			return nil, err
		}
		return acc, nil
	}
}

func (s *server) SubscriptionGardenProvider() awsclient.GardenClientProvider[subscriptionclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (subscriptionclient.AwsStsClient, error) {
		acc, err := s.Login(key, secret)
		if err != nil {
			return nil, err
		}
		return acc, nil
	}
}

func (s *server) IpRangeSkrProvider() awsclient.SkrClientProvider[awsiprangeclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsiprangeclient.Client, error) {
		acc := s.GetAccount(account)
		if acc == nil {
			return nil, ErrNoAccount
		}
		return acc.Region(region), nil
	}
}

func (s *server) NfsInstanceSkrProvider() awsclient.SkrClientProvider[awsnfsinstanceclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsnfsinstanceclient.Client, error) {
		acc := s.GetAccount(account)
		if acc == nil {
			return nil, ErrNoAccount
		}
		return acc.Region(region), nil
	}
}

func (s *server) VpcPeeringSkrProvider() awsclient.SkrClientProvider[awsvpcpeeringclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsvpcpeeringclient.Client, error) {
		acc := s.GetAccount(account)
		if acc == nil {
			return nil, ErrNoAccount
		}
		return acc.Region(region), nil
	}
}

func (s *server) ElastiCacheProviderFake() awsclient.SkrClientProvider[awsclient.ElastiCacheClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsclient.ElastiCacheClient, error) {
		acc := s.GetAccount(account)
		if acc == nil {
			return nil, ErrNoAccount
		}
		return acc.Region(region), nil
	}
}

func (s *server) ExposedDataProvider() awsclient.SkrClientProvider[awsexposeddataclient.Client] {
	return func(_ context.Context, account, region, key, secret, role string) (awsexposeddataclient.Client, error) {
		acc := s.GetAccount(account)
		if acc == nil {
			return nil, ErrNoAccount
		}
		return acc.Region(region), nil
	}
}

// TODO: shouldn't be used ==============================

func (s *server) XXX_MockConfigs(account, region string) AccountRegion {
	panic("should not be called")
	//return s.getAccountRegionContext(account, region)
}

func (s *server) xxx_getAccountRegionContext(account, region string) *accountRegionStore {
	s.m.Lock()
	defer s.m.Unlock()

	key := fmt.Sprintf("%s:%s", account, region)
	acc, ok := s.accountStores[key]
	if !ok {
		acc = newAccountRegionStore(account, region)
		s.accountStores[key] = acc
	}

	return acc
}
