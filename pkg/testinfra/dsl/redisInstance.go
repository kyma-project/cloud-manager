package dsl

import (
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
