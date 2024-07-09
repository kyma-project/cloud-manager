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

func WithGcpRedisInstanceTier(tier string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.Tier = tier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceTier", obj))
		},
	}
}

func WithGcpRedisInstanceMemorySizeGb(memorySizeGb int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.MemorySizeGb = memorySizeGb
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceMemorySizeGb", obj))
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

func WithGcpRedisInstanceTransitEncryptionMode(transitEncryptionMode string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				gcpRedisInstance.Spec.TransitEncryptionMode = transitEncryptionMode
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpRedisInstanceTransitEncryptionMode", obj))
		},
	}
}

func WithGcpRedisInstanceRedisConfigs(redisConfigs cloudresourcesv1beta1.RedisInstanceGcpConfigs) ObjAction {
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

func WithGcpRedisInstanceAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if gcpRedisInstance.Spec.AuthSecret == nil {
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
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
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
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
					gcpRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
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
