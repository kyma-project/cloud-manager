package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithRedisInstancePrimaryEndpoint(primaryEndpoint string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if redisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				redisInstance.Status.PrimaryEndpoint = primaryEndpoint
			}
		},
	}
}

func WithRedisInstanceReadEndpoint(readEndpoint string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if redisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				redisInstance.Status.ReadEndpoint = readEndpoint
			}
		},
	}
}

func WithRedisInstanceAuthString(authString string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if redisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				redisInstance.Status.AuthString = authString
			}
		},
	}
}

func HavingRedisInstanceStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.RedisInstance)
		if !ok {
			return fmt.Errorf("the object %T is not KCP RedisInstance", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the KCP RedisInstance .status.id not set")
		}
		return nil
	}
}

func CreateRedisInstance(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.RedisInstance, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.RedisInstance{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP RedisInstance must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithRedisInstanceGcp() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Gcp = &cloudcontrolv1beta1.RedisInstanceGcp{}
			}
		},
	}
}

func WithRedisInstanceAws() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Aws = &cloudcontrolv1beta1.RedisInstanceAws{}
			}
		},
	}
}

func WithKcpGcpRedisInstanceTier(tier string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.Tier = tier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceTier", obj))
		},
	}
}

func WithKcpGcpRedisInstanceMemorySizeGb(memorySizeGb int32) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.MemorySizeGb = memorySizeGb
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceMemorySizeGb", obj))
		},
	}
}

func WithKcpGcpRedisInstanceRedisVersion(redisVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.RedisVersion = redisVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceRedisVersion", obj))
		},
	}
}
