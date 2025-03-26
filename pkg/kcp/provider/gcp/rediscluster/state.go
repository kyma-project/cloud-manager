package rediscluster

import (
	"context"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"

	gcpClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	types.State

	gcpRedisCluster   *clusterpb.Cluster
	memorystoreClient client.MemorystoreClusterClient

	updateMask []string
}

type StateFactory interface {
	NewState(ctx context.Context, redisClusterState types.State) (*State, error)
}

type stateFactory struct {
	memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClusterClient]
	env                       abstractions.Environment
}

func NewStateFactory(memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClusterClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		memorystoreClientProvider: memorystoreClientProvider,
		env:                       env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, redisClusterState types.State) (*State, error) {

	memorystoreClient, err := statefactory.memorystoreClientProvider(
		ctx,
		statefactory.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return newState(redisClusterState, memorystoreClient), nil
}

func newState(redisClusterState types.State, memorystoreClient client.MemorystoreClusterClient) *State {
	return &State{
		State:             redisClusterState,
		memorystoreClient: memorystoreClient,
		updateMask:        []string{},
	}
}

func (s *State) GetRemoteRedisName() string {
	return s.Obj().GetName()
}

func (s *State) ShouldUpdateRedisCluster() bool {
	return len(s.updateMask) > 0
}

func (s *State) UpdateRedisConfigs(redisConfigs map[string]string) {
	s.updateMask = append(s.updateMask, "redis_configs") // it is 'redis_configs', GCP API says 'redisConfig', but it is wrongly documented
	s.gcpRedisCluster.RedisConfigs = redisConfigs
}
