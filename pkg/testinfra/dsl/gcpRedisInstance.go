package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateGcpRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpRedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpRedisInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpRedisInstance must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithGcpRedisInstanceDefaultSpec() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.RedisTier = cloudresourcesv1beta1.GcpRedisTierP3
				gcpRedisInstance.Spec.RedisVersion = "REDIS_7_0"
				gcpRedisInstance.Spec.AuthEnabled = true
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceDefaultSpec", obj))
		},
	}
}

func WithGcpRedisInstanceRedisTier(redisTier cloudresourcesv1beta1.GcpRedisTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.RedisTier = redisTier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceRedisTier", obj))
		},
	}
}

func WithGcpRedisInstanceRedisVersion(redisVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.RedisVersion = redisVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceRedisVersion", obj))
		},
	}
}

func WithGcpRedisInstanceAuthEnabled(authEnabled bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.AuthEnabled = authEnabled
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceAuthEnabled", obj))
		},
	}
}

func WithGcpRedisInstanceRedisConfigs(redisConfigs map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.RedisConfigs = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceRedisConfigs", obj))
		},
	}
}

func WithGcpRedisInstanceMaintenancePolicy(maintenancePolicy *cloudresourcesv1beta1.MaintenancePolicy) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.MaintenancePolicy = maintenancePolicy
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceMaintenancePolicy", obj))
		},
	}
}

func WithGcpRedisInstanceAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if gcpRedisInstance.Spec.AuthSecret == nil {
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				gcpRedisInstance.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceAuthSecretName", obj))
		},
	}
}

func WithGcpRedisInstanceAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if gcpRedisInstance.Spec.AuthSecret == nil {
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisInstance.Spec.AuthSecret.Labels == nil {
					gcpRedisInstance.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					gcpRedisInstance.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceAuthSecretLabels", obj))
		},
	}
}

func WithGcpRedisInstanceAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if gcpRedisInstance.Spec.AuthSecret == nil {
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisInstance.Spec.AuthSecret.Annotations == nil {
					gcpRedisInstance.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					gcpRedisInstance.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceAuthSecretAnnotations", obj))
		},
	}
}

func WithGcpRedisInstanceAuthSecretExtraData(extraData map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if gcpRedisInstance.Spec.AuthSecret == nil {
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if gcpRedisInstance.Spec.AuthSecret.ExtraData == nil {
					gcpRedisInstance.Spec.AuthSecret.ExtraData = map[string]string{}
				}
				for k, v := range extraData {
					gcpRedisInstance.Spec.AuthSecret.ExtraData[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceAuthSecretExtraData", obj))
		},
	}
}

func HavingGcpRedisInstanceStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpRedisInstance", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR GcpRedisInstance ID not set")
		}
		return nil
	}
}

func HavingGcpRedisInstanceStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpRedisInstance", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR GcpRedisInstance State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}

func UpdateGcpRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpRedisInstance, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)
	return clnt.Update(ctx, obj)
}
