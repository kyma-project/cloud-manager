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

	//gcp config
	gcpConfig *gcpClient.GcpConfig
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
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
	}
}

func (s *State) GetRemoteRedisName() string {
	return s.Obj().GetName()
}
