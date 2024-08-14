package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsRedisInstance(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsRedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AwsRedisInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AwsRedisInstance must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithAwsRedisInstanceDefautSpecs() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.CacheNodeType = "cache.m5.large"
				awsRedisInstance.Spec.EngineVersion = "6.x"
				awsRedisInstance.Spec.AutoMinorVersionUpgrade = true
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceDefautSpecs", obj))
		},
	}
}

func WithAwsRedisInstanceCacheNodeType(cacheNodeType string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.CacheNodeType = cacheNodeType
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceCacheNodeType", obj))
		},
	}
}

func WithAwsRedisInstanceEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceEngineVersion", obj))
		},
	}
}

func WithAwsRedisInstanceAutoMinorVersionUpgrade(autoMinorVersionUpgrade bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.AutoMinorVersionUpgrade = autoMinorVersionUpgrade
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceAutoMinorVersionUpgrade", obj))
		},
	}
}

func WithAwsRedisInstanceTransitEncryptionEnabled(transitEncryptionEnabled bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.TransitEncryptionEnabled = transitEncryptionEnabled
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceTransitEncryptionEnabled", obj))
		},
	}
}

func WithAwsRedisInstanceParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				awsRedisInstance.Spec.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceParameters", obj))
		},
	}
}

func WithAwsRedisInstanceAuthSecretName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				if awsRedisInstance.Spec.AuthSecret == nil {
					awsRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
				}
				awsRedisInstance.Spec.AuthSecret.Name = name
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceAuthSecretName", obj))
		},
	}
}

func WithAwsRedisInstanceAuthSecretLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				if awsRedisInstance.Spec.AuthSecret == nil {
					awsRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
				}
				if awsRedisInstance.Spec.AuthSecret.Labels == nil {
					awsRedisInstance.Spec.AuthSecret.Labels = map[string]string{}
				}
				for k, v := range labels {
					awsRedisInstance.Spec.AuthSecret.Labels[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceAuthSecretLabels", obj))
		},
	}
}

func WithAwsRedisInstanceAuthSecretAnnotations(annotations map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if awsRedisInstance, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				if awsRedisInstance.Spec.AuthSecret == nil {
					awsRedisInstance.Spec.AuthSecret = &cloudresourcesv1beta1.AuthSecretSpec{}
				}
				if awsRedisInstance.Spec.AuthSecret.Annotations == nil {
					awsRedisInstance.Spec.AuthSecret.Annotations = map[string]string{}
				}
				for k, v := range annotations {
					awsRedisInstance.Spec.AuthSecret.Annotations[k] = v
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithAwsRedisInstanceAuthSecretAnnotations", obj))
		},
	}
}

func HavingAwsRedisInstanceStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsRedisInstance", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR AwsRedisInstance ID not set")
		}
		return nil
	}
}

func HavingAwsRedisInstanceStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not SKR AwsRedisInstance", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR AwsRedisInstance State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}
