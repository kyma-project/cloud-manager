package rediscluster

import (
	"context"
	"fmt"

	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/rediscluster/client"
	instanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
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

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud rediscluster client: %w", err)
	}

	return &State{
		State:  redisClusterState,
		client: c,
	}, nil
}
