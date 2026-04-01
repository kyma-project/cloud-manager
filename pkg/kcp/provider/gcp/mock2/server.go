package mock2

import (
	"fmt"
	"sync"

	"github.com/elliotchance/pie/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	gcpvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcnetwork/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func New() Server {
	return &server{}
}

type server struct {
	m sync.Mutex

	subscriptions []Store
}

// Providers START - add new provider methods below ====================================================

func (s *server) ExposedDataProvider() gcpclient.GcpClientProvider[gcpexposeddataclient.Client] {
	return func(projectId string) gcpexposeddataclient.Client {
		return s.GetSubscription(projectId)
	}
}

func (s *server) VpcNetworkProvider() gcpclient.GcpClientProvider[gcpvpcnetworkclient.Client] {
	return func(projectId string) gcpvpcnetworkclient.Client {
		return s.GetSubscription(projectId)
	}
}

func (s *server) RedisInstanceProvider() gcpclient.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient] {
	return func(projectId string) gcpredisinstanceclient.MemorystoreClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpredisinstanceclient.NewMemorystoreClientFromRedisInstanceClient(sub)
	}
}

func (s *server) SubnetComputeProvider() gcpclient.GcpClientProvider[gcpsubnetclient.ComputeClient] {
	return func(projectId string) gcpsubnetclient.ComputeClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpsubnetclient.NewComputeClientFromSubnetClient(sub)
	}
}

func (s *server) SubnetNetworkConnectivityProvider() gcpclient.GcpClientProvider[gcpsubnetclient.NetworkConnectivityClient] {
	return func(projectId string) gcpsubnetclient.NetworkConnectivityClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpsubnetclient.NewNetworkConnectivityClientFromWrapped(sub)
	}
}

func (s *server) SubnetRegionOperationsProvider() gcpclient.GcpClientProvider[gcpsubnetclient.RegionOperationsClient] {
	return func(projectId string) gcpsubnetclient.RegionOperationsClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpsubnetclient.NewRegionOperationsClientFromWrapped(sub)
	}
}

func (s *server) RedisClusterProvider() gcpclient.GcpClientProvider[gcpredisclusterclient.MemorystoreClusterClient] {
	return func(projectId string) gcpredisclusterclient.MemorystoreClusterClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpredisclusterclient.NewMemorystoreClientFromRedisClusterClient(sub)
	}
}

func (s *server) NfsInstanceV2Provider() gcpclient.GcpClientProvider[gcpnfsinstancev2client.FilestoreClient] {
	return func(projectId string) gcpnfsinstancev2client.FilestoreClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpnfsinstancev2client.NewFilestoreClientFromFilestoreClient(sub)
	}
}

func (s *server) NfsBackupV2Provider() gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient] {
	return func(projectId string) gcpnfsbackupclientv2.FileBackupClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpnfsbackupclientv2.NewFileBackupClientFromFilestoreClient(sub)
	}
}

// Providers END - add new provider methods above ===========================================

func (s *server) NewSubscription(prefix string) Store {
	s.m.Lock()
	defer s.m.Unlock()

	if prefix == "" {
		prefix = "e2e"
	}
	name := fmt.Sprintf("%s-%s", prefix, util.RandomString(10))

	sub := newStore(name, s)
	s.subscriptions = append(s.subscriptions, sub)

	return sub
}

// GetSubscription returns previously created subscription with the given projectId. If there is no subscription with
// such projectId, nil is returned, intentionally so that reconciler fails, which would indicate invalid test setup
// and a signal to developer to create the subscription at the beginning of the test.
func (s *server) GetSubscription(projectId string) Store {
	s.m.Lock()
	defer s.m.Unlock()

	for _, p := range s.subscriptions {
		if p.ProjectId() == projectId {
			return p
		}
	}
	return nil
}

func (s *server) DeleteSubscription(projectId string) {
	s.m.Lock()
	defer s.m.Unlock()

	s.subscriptions = pie.FilterNot(s.subscriptions, func(s Store) bool {
		return s.ProjectId() == projectId
	})
}

var _ Server = (*server)(nil)
