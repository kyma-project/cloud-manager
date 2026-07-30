package dsl

import (
	"context"
	"errors"
	"fmt"
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ======================== KCP RedisInstance (Alicloud) ========================

func WithRedisInstanceAlicloud() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Alicloud = &cloudcontrolv1beta1.RedisInstanceAlicloud{}
			}
		},
	}
}

func WithKcpAlicloudRedisInstanceClass(instanceClass string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Alicloud.InstanceClass = instanceClass
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisInstanceClass", obj))
		},
	}
}

func WithKcpAlicloudRedisEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Alicloud.EngineVersion = engineVersion
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisEngineVersion", obj))
		},
	}
}

func WithKcpAlicloudRedisReadOnlyCount(readOnlyCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Alicloud.ReadOnlyCount = readOnlyCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisReadOnlyCount", obj))
		},
	}
}

func WithKcpAlicloudRedisParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Alicloud.Parameters = parameters
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisParameters", obj))
		},
	}
}

// ======================== KCP RedisCluster (Alicloud) ========================

func WithRedisClusterAlicloud() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud = &cloudcontrolv1beta1.RedisClusterAlicloud{}
			}
		},
	}
}

func WithKcpAlicloudRedisClusterInstanceClass(instanceClass string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud.InstanceClass = instanceClass
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisClusterInstanceClass", obj))
		},
	}
}

func WithKcpAlicloudRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisClusterShardCount", obj))
		},
	}
}

func WithKcpAlicloudRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				x.Spec.Instance.Alicloud.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAlicloudRedisClusterReplicasPerShard", obj))
		},
	}
}

// ======================== SKR AlicloudRedisInstance ========================

func CreateAlicloudRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AlicloudRedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AlicloudRedisInstance{}
	}
	NewObjActions(opts...).
		Append(WithNamespace(DefaultSkrNamespace)).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AlicloudRedisInstance must have name set")
	}

	return clnt.Create(ctx, obj)
}

func WithAlicloudRedisInstanceRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				x.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceRedisTier", obj))
		},
	}
}

func WithAlicloudRedisInstanceEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				x.Spec.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceEngineVersion", obj))
		},
	}
}

func WithAlicloudRedisInstanceParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				x.Spec.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceParameters", obj))
		},
	}
}

func WithAlicloudRedisInstanceAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				x.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceAuthSecretName", obj))
		},
	}
}

func WithAlicloudRedisInstanceAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.Labels == nil {
					x.Spec.AuthSecret.Labels = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.Labels, labels)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceAuthSecretLabels", obj))
		},
	}
}

func WithAlicloudRedisInstanceAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.Annotations == nil {
					x.Spec.AuthSecret.Annotations = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.Annotations, annotations)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceAuthSecretAnnotations", obj))
		},
	}
}

func WithAlicloudRedisInstanceAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisInstance); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.ExtraData == nil {
					x.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.ExtraData, extraData)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisInstanceAuthSecretExtraData", obj))
		},
	}
}

// ======================== SKR AlicloudRedisCluster ========================

func CreateAlicloudRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AlicloudRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AlicloudRedisCluster{}
	}
	NewObjActions(opts...).
		Append(WithNamespace(DefaultSkrNamespace)).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AlicloudRedisCluster must have name set")
	}

	return clnt.Create(ctx, obj)
}

func WithAlicloudRedisClusterRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisClusterTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				x.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterRedisTier", obj))
		},
	}
}

func WithAlicloudRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				x.Spec.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterShardCount", obj))
		},
	}
}

func WithAlicloudRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				x.Spec.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterReplicasPerShard", obj))
		},
	}
}

func WithAlicloudRedisClusterEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				x.Spec.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterEngineVersion", obj))
		},
	}
}

func WithAlicloudRedisClusterParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				x.Spec.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterParameters", obj))
		},
	}
}

func WithAlicloudRedisClusterAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				x.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterAuthSecretName", obj))
		},
	}
}

func WithAlicloudRedisClusterAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.Labels == nil {
					x.Spec.AuthSecret.Labels = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.Labels, labels)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterAuthSecretLabels", obj))
		},
	}
}

func WithAlicloudRedisClusterAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.Annotations == nil {
					x.Spec.AuthSecret.Annotations = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.Annotations, annotations)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterAuthSecretAnnotations", obj))
		},
	}
}

func WithAlicloudRedisClusterAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AlicloudRedisCluster); ok {
				if x.Spec.AuthSecret == nil {
					x.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if x.Spec.AuthSecret.ExtraData == nil {
					x.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				maps.Copy(x.Spec.AuthSecret.ExtraData, extraData)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAlicloudRedisClusterAuthSecretExtraData", obj))
		},
	}
}
