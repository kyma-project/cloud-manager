package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsRedisCluster, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsRedisCluster{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsRedisCluster must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithAwsRedisClusterDefautSpecs() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.RedisTier = cloudresourcesv1beta1.AwsRedisTierC4
				awsRedisCluster.Spec.EngineVersion = "6.x"
				awsRedisCluster.Spec.AutoMinorVersionUpgrade = true
				awsRedisCluster.Spec.AuthEnabled = false
				awsRedisCluster.Spec.ShardCount = 3
				awsRedisCluster.Spec.ReplicasPerShard = 1
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterDefautSpecs", obj))
		},
	}
}

func WithAwsRedisClusterRedisTier(redisTier cloudresourcesv1beta1.AwsRedisClusterTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterRedisTier", obj))
		},
	}
}

func WithAwsRedisClusterShardCount(shardCount int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.ShardCount = shardCount
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterShardCount", obj))
		},
	}
}

func WithAwsRedisClusterReplicasPerShard(replicasPerShard int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.ReplicasPerShard = replicasPerShard
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterReplicasPerShard", obj))
		},
	}
}

func WithAwsRedisClusterEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterEngineVersion", obj))
		},
	}
}

func WithAwsRedisClusterAutoMinorVersionUpgrade(autoMinorVersionUpgrade bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.AutoMinorVersionUpgrade = autoMinorVersionUpgrade
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAutoMinorVersionUpgrade", obj))
		},
	}
}

func WithAwsRedisClusterAuthEnabled(authEnabled bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.AuthEnabled = authEnabled
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAuthEnabled", obj))
		},
	}
}

func WithAwsRedisClusterPreferredMaintenanceWindow(preferredMaintenanceWindow *string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.PreferredMaintenanceWindow = preferredMaintenanceWindow
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterPreferredMaintenanceWindow", obj))
		},
	}
}

func WithAwsRedisClusterParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				awsRedisCluster.Spec.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterParameters", obj))
		},
	}
}

func WithAwsRedisClusterAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				if awsRedisCluster.Spec.AuthSecret == nil {
					awsRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				awsRedisCluster.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAuthSecretName", obj))
		},
	}
}

func WithAwsRedisClusterAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				if awsRedisCluster.Spec.AuthSecret == nil {
					awsRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if awsRedisCluster.Spec.AuthSecret.Labels == nil {
					awsRedisCluster.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					awsRedisCluster.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAuthSecretLabels", obj))
		},
	}
}

func WithAwsRedisClusterAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				if awsRedisCluster.Spec.AuthSecret == nil {
					awsRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if awsRedisCluster.Spec.AuthSecret.Annotations == nil {
					awsRedisCluster.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					awsRedisCluster.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAuthSecretAnnotations", obj))
		},
	}
}

func WithAwsRedisClusterAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisCluster, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster); ok {
				if awsRedisCluster.Spec.AuthSecret == nil {
					awsRedisCluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if awsRedisCluster.Spec.AuthSecret.ExtraData == nil {
					awsRedisCluster.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				for k, v := range extraData {
					awsRedisCluster.Spec.AuthSecret.ExtraData[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisClusterAuthSecretExtraData", obj))
		},
	}
}

func HavingAwsRedisClusterStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsRedisCluster", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AwsRedisCluster ID not set")
		}
		return nil
	}
}

func HavingAwsRedisClusterStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsRedisCluster)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsRedisCluster", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR AwsRedisCluster State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}

func UpdateAwsRedisCluster(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsRedisCluster, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)
	return clnt.Update(ctx, obj)
}
