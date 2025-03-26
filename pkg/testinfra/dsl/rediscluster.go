package dsl

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithRedisClusterGcp() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Gcp = &cloudcontrolv1beta1.RedisClusterGcp{}
			}
		},
	}
}

func WithKcpGcpRedisClusterConfigs(redisConfigs map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				gcpRedisCluster.Spec.Instance.Gcp.RedisConfigs = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterConfigs", obj))
		},
	}
}

func WithKcpGcpRedisClusterNodeType(nodeType string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				gcpRedisCluster.Spec.Instance.Gcp.NodeType = nodeType
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterNodeType", obj))
		},
	}
}

func WithKcpGcpRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				gcpRedisCluster.Spec.Instance.Gcp.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterShardCount", obj))
		},
	}
}

func WithKcpGcpRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				gcpRedisCluster.Spec.Instance.Gcp.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisClusterReplicasPerShard", obj))
		},
	}
}
