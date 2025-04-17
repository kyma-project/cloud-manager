package rediscluster

import (
	"context"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	focal.State

	subnet *cloudcontrolv1beta1.GcpSubnet

	gcpRedisCluster   *clusterpb.Cluster
	memorystoreClient client.MemorystoreClusterClient

	caCerts string

	updateMask []string
}

type StateFactory interface {
	NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
	memorystoreClientProvider gcpclient.ClientProvider[client.MemorystoreClusterClient]
	env                       abstractions.Environment
}

func NewStateFactory(
	memorystoreClientProvider gcpclient.ClientProvider[client.MemorystoreClusterClient],
	env abstractions.Environment,
) StateFactory {
	return &stateFactory{
		memorystoreClientProvider: memorystoreClientProvider,
		env:                       env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {

	memorystoreClient, err := statefactory.memorystoreClientProvider(
		ctx,
		statefactory.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	return newState(focalState, memorystoreClient), nil
}

func newState(focalState focal.State, memorystoreClient client.MemorystoreClusterClient) *State {
	return &State{
		State:             focalState,
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

func (s *State) ObjAsGcpRedisCluster() *cloudcontrolv1beta1.GcpRedisCluster {
	return s.Obj().(*cloudcontrolv1beta1.GcpRedisCluster)
}

func (s *State) Subnet() *cloudcontrolv1beta1.GcpSubnet {
	return s.subnet
}

func (s *State) SetSubnet(subnet *cloudcontrolv1beta1.GcpSubnet) {
	s.subnet = subnet
}
