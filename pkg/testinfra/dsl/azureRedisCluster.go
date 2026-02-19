package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRedisCluster{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRedisCluster must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithAzureRedisClusterRedisVersion(redisVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.RedisVersion = redisVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterRedisVersion", obj))
		},
	}
}

func WithAzureRedisClusterRedisTier(redisTier cloudresourcesv1beta1.AzureRedisClusterTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterRedisTire", obj))
		},
	}
}

func WithAzureRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterShardCount", obj))
		},
	}
}

func WithAzureRedisClusterReplicaCount(replicaCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.ReplicasPerPrimary = replicaCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterReplicaCount", obj))
		},
	}
}

func WithAzureRedisClusterRedisConfigs(redisConfigs cloudresourcesv1beta1.RedisClusterAzureConfigs) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.RedisConfiguration = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterRedisConfigs", obj))
		},
	}
}

func WithAzureRedisClusterAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				if azureRedisCluster.Spec.AuthSecret == nil {
					azureRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				azureRedisCluster.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterAuthSecretName", obj))
		},
	}
}

func WithAzureRedisClusterAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				if azureRedisCluster.Spec.AuthSecret == nil {
					azureRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if azureRedisCluster.Spec.AuthSecret.Labels == nil {
					azureRedisCluster.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					azureRedisCluster.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterAuthSecretLabels", obj))
		},
	}
}

func WithAzureRedisClusterAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				if azureRedisCluster.Spec.AuthSecret == nil {
					azureRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if azureRedisCluster.Spec.AuthSecret.Annotations == nil {
					azureRedisCluster.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					azureRedisCluster.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterAuthSecretAnnotations", obj))
		},
	}
}

func WithAzureRedisClusterAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				if azureRedisCluster.Spec.AuthSecret == nil {
					azureRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if azureRedisCluster.Spec.AuthSecret.ExtraData == nil {
					azureRedisCluster.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				for k, v := range extraData {
					azureRedisCluster.Spec.AuthSecret.ExtraData[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterAuthSecretExtraData", obj))
		},
	}
}

func WithAzureRedisClusterVolume(volume *cloudresourcesv1beta1.RedisAuthSecretSpec) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.Volume = volume
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterVolume", obj))
		},
	}
}

func HavingAzureRedisClusterStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AzureRedisCluster", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AzureRedisCluster ID not set")
		}
		return nil
	}
}

func HavingAzureRedisClusterStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AzureRedisCluster", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR AzureRedisCluster State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}

func UpdateAzureRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRedisCluster{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRedisCluster must have name set")
	}

	err := clnt.Update(ctx, obj)
	return err
}

func WithAzureRedisClusterDefaultSpecs() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisCluster, ok := obj.(*cloudresourcesv1beta1.AzureRedisCluster); ok {
				azureRedisCluster.Spec.RedisTier = cloudresourcesv1beta1.AzureRedisTierC4
				azureRedisCluster.Spec.RedisVersion = "6"
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisClusterDefaultSpecs", obj))
		},
	}
}
