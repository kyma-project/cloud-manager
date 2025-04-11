package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateGcpRedisCluster(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.GcpRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.GcpRedisCluster{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP GcpRedisCluster must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithKcpGcpRedisClusterConfigs(redisConfigs map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.RedisConfigs = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterConfigs", obj))
		},
	}
}

func WithKcpGcpRedisClusterNodeType(nodeType string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.NodeType = nodeType
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterNodeType", obj))
		},
	}
}

func WithKcpGcpRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterShardCount", obj))
		},
	}
}

func WithKcpGcpRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterReplicasPerShard", obj))
		},
	}
}

func HavingGcpRedisClusterStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not KCP GcpRedisCluster", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the KCP GcpRedisCluster .status.id not set")
		}
		return nil
	}
}

func WithGcpRedisClusterDiscoveryEndpoint(discoveryEndpoint string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if redisCluster, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				redisCluster.Status.DiscoveryEndpoint = discoveryEndpoint
			}
		},
	}
}
