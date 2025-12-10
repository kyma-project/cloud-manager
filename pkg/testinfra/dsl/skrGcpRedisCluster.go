package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateSkrGcpRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpRedisCluster{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpRedisCluster must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithSkrGcpRedisClusterDefaultSpec() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.RedisTier = cloudresourcesv1beta1.GcpRedisClusterTierC3
				gcpRedisCluster.Spec.ShardCount = 5
				gcpRedisCluster.Spec.ReplicasPerShard = 2
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterDefaultSpec", obj))
		},
	}
}

func WithSkrGcpRedisClusterRedisTier(redisTier cloudresourcesv1beta1.GcpRedisClusterTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterRedisTier", obj))
		},
	}
}

func WithSkrGcpRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterShardCount", obj))
		},
	}
}

func WithSkrGcpRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterReplicasPerShard", obj))
		},
	}
}

func WithSkrGcpRedisClusterRedisConfigs(redisConfigs map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				gcpRedisCluster.Spec.RedisConfigs = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterRedisConfigs", obj))
		},
	}
}

func WithSkrGcpRedisClusterAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				if gcpRedisCluster.Spec.AuthSecret == nil {
					gcpRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				gcpRedisCluster.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterAuthSecretName", obj))
		},
	}
}

func WithSkrGcpRedisClusterAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				if gcpRedisCluster.Spec.AuthSecret == nil {
					gcpRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisCluster.Spec.AuthSecret.Labels == nil {
					gcpRedisCluster.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					gcpRedisCluster.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterAuthSecretLabels", obj))
		},
	}
}

func WithSkrGcpRedisClusterAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				if gcpRedisCluster.Spec.AuthSecret == nil {
					gcpRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisCluster.Spec.AuthSecret.Annotations == nil {
					gcpRedisCluster.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					gcpRedisCluster.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterAuthSecretAnnotations", obj))
		},
	}
}

func WithSkrGcpRedisClusterAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisCluster, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				if gcpRedisCluster.Spec.AuthSecret == nil {
					gcpRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisCluster.Spec.AuthSecret.ExtraData == nil {
					gcpRedisCluster.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				for k, v := range extraData {
					gcpRedisCluster.Spec.AuthSecret.ExtraData[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpRedisClusterAuthSecretExtraData", obj))
		},
	}
}

func HavingSkrGcpRedisClusterStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpRedisCluster", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR GcpRedisCluster ID not set")
		}
		return nil
	}
}

func HavingSkrGcpRedisClusterStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpRedisCluster", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR GcpRedisCluster State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}

func UpdateGcpRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpRedisCluster, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)
	return clnt.Update(ctx, obj)
}
