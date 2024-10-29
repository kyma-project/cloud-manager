package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRedisInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRedisInstance must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithAzureRedisInstanceRedisVersion(redisVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				azureRedisInstance.Spec.RedisVersion = redisVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceRedisVersion", obj))
		},
	}
}

func WithAzureRedisInstanceSKUCapacity(sku cloudresourcesv1beta1.AzureRedisSKU) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				azureRedisInstance.Spec.SKU = sku
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceSKUCapacity", obj))
		},
	}
}

func WithAzureRedisInstanceRedisConfigs(redisConfigs cloudresourcesv1beta1.RedisInstanceAzureConfigs) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				azureRedisInstance.Spec.RedisConfiguration = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceRedisConfigs", obj))
		},
	}
}

func WithAzureRedisInstanceAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				if azureRedisInstance.Spec.AuthSecret == nil {
					azureRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				azureRedisInstance.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceAuthSecretName", obj))
		},
	}
}

func WithAzureRedisInstanceAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				if azureRedisInstance.Spec.AuthSecret == nil {
					azureRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if azureRedisInstance.Spec.AuthSecret.Labels == nil {
					azureRedisInstance.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					azureRedisInstance.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceAuthSecretLabels", obj))
		},
	}
}

func WithAzureRedisInstanceAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				if azureRedisInstance.Spec.AuthSecret == nil {
					azureRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
				}
				if azureRedisInstance.Spec.AuthSecret.Annotations == nil {
					azureRedisInstance.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					azureRedisInstance.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceAuthSecretAnnotations", obj))
		},
	}
}

func HavingAzureRedisInstanceStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AzureRedisInstance", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AzureRedisInstance ID not set")
		}
		return nil
	}
}

func HavingAzureRedisInstanceStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AzureRedisInstance", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR AzureRedisInstance State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}

func UpdateAzureRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureRedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureRedisInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureRedisInstance must have name set")
	}

	err := clnt.Update(ctx, obj)
	return err
}

func WithAzureRedisInstanceDefautSpecs() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				azureRedisInstance.Spec.ShardCount = 1
				azureRedisInstance.Spec.SKU.Capacity = 1
				azureRedisInstance.Spec.RedisVersion = "7"
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAzureRedisInstanceDefautSpecs", obj))
		},
	}
}
