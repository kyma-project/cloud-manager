// Package client wraps the AliCloud r-kvstore SDK for sharded cloud-native
// cluster operations. It extends the standard-instance Client interface from
// pkg/kcp/provider/alicloud/redisinstance/client with the sharding operations
// needed by the AlicloudRedisCluster reconciler.
//
// Per issue #2012 design decision 8, InstanceClass changes and ShardCount
// changes cannot be combined in a single ModifyInstanceSpec call. The
// reconciler therefore uses:
//
//   - ModifyInstanceSpec (from the embedded Client) for InstanceClass
//     and ReplicasPerShard changes
//   - AddShardingNode / DeleteShardingNode (on this interface) for
//     ShardCount changes
package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rkvstore "github.com/alibabacloud-go/r-kvstore-20150101/v7/client"

	instanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
)

// Client extends the standard-instance Client with cluster-only sharding
// operations. Both AddShardingNode and DeleteShardingNode accept the target
// ShardCount (absolute, not delta) - the SDK computes the required
// add/remove count internally.
type Client interface {
	instanceclient.Client

	// AddShardingNode grows the cluster to the given target shard count.
	AddShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error

	// DeleteShardingNode shrinks the cluster to the given target shard count.
	DeleteShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error
}

// ClientProvider mirrors the credential/region-scoped constructor signature
// used by all AliCloud client packages in cloud-manager.
type ClientProvider func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error)

// NewClientProvider returns a ClientProvider backed by the real AliCloud
// r-kvstore SDK. A single *rkvstore.Client is constructed and shared between
// the embedded instance operations and the cluster-only sharding operations,
// so exactly one SDK connection is opened per reconcile.
func NewClientProvider() ClientProvider {
	return func(ctx context.Context, region, accessKeyId, accessKeySecret string) (Client, error) {
		config := &openapi.Config{
			AccessKeyId:     new(accessKeyId),
			AccessKeySecret: new(accessKeySecret),
			RegionId:        new(region),
		}
		config.Endpoint = new(fmt.Sprintf("r-kvstore.%s.aliyuncs.com", region))

		rc, err := rkvstore.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud r-kvstore cluster client: %w", err)
		}

		return &alicloudRedisClusterClient{
			Client: instanceclient.NewClientFromSDK(rc, region),
			c:      rc,
			region: region,
		}, nil
	}
}

type alicloudRedisClusterClient struct {
	instanceclient.Client
	c      *rkvstore.Client
	region string
}

func (c *alicloudRedisClusterClient) AddShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	req := &rkvstore.AddShardingNodeRequest{
		InstanceId: new(instanceId),
		ShardCount: new(targetShardCount),
	}
	if _, err := c.c.AddShardingNode(req); err != nil {
		return fmt.Errorf("error adding sharding node to alicloud r-kvstore cluster %s: %w", instanceId, err)
	}
	return nil
}

func (c *alicloudRedisClusterClient) DeleteShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	req := &rkvstore.DeleteShardingNodeRequest{
		InstanceId: new(instanceId),
		ShardCount: new(targetShardCount),
	}
	if _, err := c.c.DeleteShardingNode(req); err != nil {
		return fmt.Errorf("error deleting sharding node from alicloud r-kvstore cluster %s: %w", instanceId, err)
	}
	return nil
}
