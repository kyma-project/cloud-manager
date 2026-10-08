package rediscluster

import (
	"context"
	"fmt"

	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudmetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/metrics"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	instanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	alicloudutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

type State struct {
	types.State

	client   alicloudclient.Client
	instance *instanceclient.InstanceInfo
}

type StateFactory interface {
	NewState(ctx context.Context, redisClusterState types.State) (*State, error)
}

func NewStateFactory(clientProvider alicloudclient.ClientProvider) StateFactory {
	return &stateFactory{clientProvider: clientProvider}
}

type stateFactory struct {
	clientProvider alicloudclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, redisClusterState types.State) (*State, error) {
	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret
	region := redisClusterState.Scope().Spec.Region

	accountId := alicloudmetrics.AccountIdFromScope(redisClusterState.Scope())
	if accountId == "" {
		return nil, fmt.Errorf("scope %q for AliCloud RedisCluster has no account id", redisClusterState.Scope().Name)
	}
	ctx = alicloudmetrics.AccountIdIntoContext(ctx, accountId)

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret, alicloudutil.RoleArnDefault(accountId))
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud rediscluster client: %w", err)
	}

	return &State{
		State:  redisClusterState,
		client: c,
	}, nil
}
