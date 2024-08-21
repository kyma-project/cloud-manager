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

	//gcp config
	gcpConfig *gcpClient.GcpConfig
}

type StateFactory interface {
	NewState(ctx context.Context, redisInstanceState types.State) (*State, error)
}

type stateFactory struct {
	memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClient]
	env                       abstractions.Environment
	gcpConfig                 *gcpClient.GcpConfig
}

func NewStateFactory(memorystoreClientProvider gcpClient.ClientProvider[client.MemorystoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		memorystoreClientProvider: memorystoreClientProvider,
		env:                       env,
		gcpConfig:                 gcpClient.GetGcpConfig(env),
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
	return newState(redisInstanceState, memorystoreClient, statefactory.gcpConfig), nil
}

func newState(redisInstanceState types.State, memorystoreClient client.MemorystoreClient, gcpConfig *gcpClient.GcpConfig) *State {
	return &State{
		State:             redisInstanceState,
		memorystoreClient: memorystoreClient,
		gcpConfig:         gcpConfig,
		updateMask:        []string{},
	}
}

func (s *State) GetRemoteRedisName() string {
	return s.Obj().GetName()
}

func (s *State) ShouldUpdateRedisInstance() bool {
	return len(s.updateMask) > 0
}

func (s *State) UpdateRedisConfigs(redisConfigs map[string]string) {
	s.updateMask = append(s.updateMask, "redis_configs") // it is 'redis_configs', GCP API says 'redisConfig', but it is wrongly documented
	s.gcpRedisInstance.RedisConfigs = redisConfigs
}

func (s *State) UpdateMemorySizeGb(memorySizeGb int32) {
	s.updateMask = append(s.updateMask, "memory_size_gb")
	s.gcpRedisInstance.MemorySizeGb = memorySizeGb
}
