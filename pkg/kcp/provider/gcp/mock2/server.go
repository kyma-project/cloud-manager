package mock2

import (
	"fmt"
	"sync"

	"github.com/elliotchance/pie/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
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

// RedisInstanceProvider cannot return s.GetSubscription() directly because MemorystoreClient
// has value-add methods (CreateRedisInstanceWithOptions, GetRedisInstanceWithAuth) that Store
// does not implement.
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
		return s.GetSubscription(projectId)
	}
}

// SubnetNetworkConnectivityProvider cannot return s.GetSubscription() directly because
// NetworkConnectivityClient has a value-add method (CreateServiceConnectionPolicyForRedis)
// that Store does not implement.
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
		return s.GetSubscription(projectId)
	}
}

// RedisClusterProvider cannot return s.GetSubscription() directly because MemorystoreClusterClient
// has value-add methods (CreateRedisClusterWithOptions, GetRedisClusterCertificateString) that
// Store does not implement.
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
		return s.GetSubscription(projectId)
	}
}

func (s *server) NfsBackupV2Provider() gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient] {
	return func(projectId string) gcpnfsbackupclientv2.FileBackupClient {
		return s.GetSubscription(projectId)
	}
}

// NfsRestoreV2Provider cannot return s.GetSubscription() directly because FileRestoreClient
// has a value-add method (FindRestoreOperation) that Store does not implement.
func (s *server) NfsRestoreV2Provider() gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient] {
	return func(projectId string) gcpnfsrestoreclientv2.FileRestoreClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpnfsrestoreclientv2.NewFileRestoreClientFromFilestoreClient(sub)
	}
}

func (s *server) IpRangeComputeProvider() gcpclient.GcpClientProvider[gcpiprangeclient.ComputeClient] {
	return func(projectId string) gcpiprangeclient.ComputeClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpiprangeclient.NewComputeClientFromWrapped(sub, sub)
	}
}

func (s *server) IpRangeServiceNetworkingProvider() gcpclient.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient] {
	return func(projectId string) gcpiprangeclient.ServiceNetworkingClient {
		sub := s.GetSubscription(projectId)
		if sub == nil {
			return nil
		}
		return gcpiprangeclient.NewServiceNetworkingClientFromWrapped(sub)
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
