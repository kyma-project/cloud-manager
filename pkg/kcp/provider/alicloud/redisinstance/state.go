package redisinstance

import (
	"context"
	"fmt"

	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

// State is the per-reconcile state for the KCP AliCloud RedisInstance
// pipeline. It embeds the shared redisinstance types.State (which exposes the
// Scope, IpRange, and KCP RedisInstance object) and augments it with:
//
//   - the AliCloud r-kvstore client, credential- and region-scoped
//   - a cached DescribeInstance snapshot (nil until loadRedis populates it)
type State struct {
	types.State

	client   alicloudclient.Client
	instance *alicloudclient.InstanceInfo
}

// StateFactory constructs per-reconcile State instances.
type StateFactory interface {
	NewState(ctx context.Context, redisInstanceState types.State) (*State, error)
}

// NewStateFactory returns a StateFactory backed by the given ClientProvider.
// In production this is alicloudclient.NewClientProvider(); in tests it is
// the mock server's RedisInstanceClientProvider().
func NewStateFactory(clientProvider alicloudclient.ClientProvider) StateFactory {
	return &stateFactory{clientProvider: clientProvider}
}

type stateFactory struct {
	clientProvider alicloudclient.ClientProvider
}

func (f *stateFactory) NewState(ctx context.Context, redisInstanceState types.State) (*State, error) {
	accessKeyId := alicloudconfig.AlicloudConfig.AccessKeyId
	accessKeySecret := alicloudconfig.AlicloudConfig.AccessKeySecret
	region := redisInstanceState.Scope().Spec.Region

	c, err := f.clientProvider(ctx, region, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("error creating alicloud redisinstance client: %w", err)
	}

	return &State{
		State:  redisInstanceState,
		client: c,
	}, nil
}
