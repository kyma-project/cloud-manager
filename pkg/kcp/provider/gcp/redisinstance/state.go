package redisinstance

import (
	"context"

	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"

	gcpClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	types.State

	gcpRedisInstance     *redispb.Instance
	gcpRedisInstanceAuth *redispb.InstanceAuthString
	memorystoreClient    client.MemorystoreClient

	updateMask []string
}

type StateFactory interface {
	NewState(ctx context.Context, redisInstanceState types.State) (*State, error)
}

type stateFactory struct {
	memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClient]
	env                       abstractions.Environment
}

func NewStateFactory(memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		memorystoreClientProvider: memorystoreClientProvider,
		env:                       env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, redisInstanceState types.State) (*State, error) {

	memorystoreClient, err := statefactory.memorystoreClientProvider(
		ctx,
		statefactory.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return newState(redisInstanceState, memorystoreClient), nil
}

func newState(redisInstanceState types.State, memorystoreClient client.MemorystoreClient) *State {
	return &State{
		State:             redisInstanceState,
		memorystoreClient: memorystoreClient,
		updateMask:        []string{},
	}
}

func (s *State) GetRemoteRedisName() string {
	return s.Obj().GetName()
}

func (s *State) ShouldUpdateRedisInstance() bool {
	return len(s.updateMask) > 0
}

func (s *State) ShouldUpgradeRedisInstance() bool {
	return s.gcpRedisInstance.RedisVersion != s.ObjAsRedisInstance().Spec.Instance.Gcp.RedisVersion
}

func (s *State) UpdateRedisConfigs(redisConfigs map[string]string) {
	s.updateMask = append(s.updateMask, "redis_configs") // it is 'redis_configs', GCP API says 'redisConfig', but it is wrongly documented
	s.gcpRedisInstance.RedisConfigs = redisConfigs
}

func (s *State) UpdateMemorySizeGb(memorySizeGb int32) {
	s.updateMask = append(s.updateMask, "memory_size_gb")
	s.gcpRedisInstance.MemorySizeGb = memorySizeGb
}

func (s *State) UpdateMaintenancePolicy(policy *redispb.MaintenancePolicy) {
	s.updateMask = append(s.updateMask, "maintenance_policy")
	s.gcpRedisInstance.MaintenancePolicy = policy
}

func (s *State) UpdateAuthEnabled(authEnabled bool) {
	s.updateMask = append(s.updateMask, "auth_enabled")
	s.gcpRedisInstance.AuthEnabled = authEnabled
}

func (s *State) UpdateReplicaCount(replicaCount int32) {
	s.updateMask = append(s.updateMask, "replica_count")
	s.gcpRedisInstance.ReplicaCount = replicaCount
}
