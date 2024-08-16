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

func WithKcpGcpRedisInstanceTransitEncryption(transitEncryption *cloudcontrolv1beta1.TransitEncryptionGcp) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.TransitEncryption = transitEncryption
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceTransitEncryption", obj))
		},
	}
}

func WithKcpGcpRedisInstanceConfigs(redisConfigs map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.RedisConfigs = redisConfigs
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceConfigs", obj))
		},
	}
}

func WithKcpGcpRedisInstanceMaintenancePolicy(maintenancePolicy *cloudcontrolv1beta1.MaintenancePolicyGcp) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy = maintenancePolicy
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpGcpRedisInstanceMaintenancePolicy", obj))
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

func WithKcpAwsCacheNodeType(cacheNodeType string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.CacheNodeType = cacheNodeType
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsCacheNodeType", obj))
		},
	}
}

func WithKcpAwsEngineVersion(engineVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.EngineVersion = engineVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsEngineVersion", obj))
		},
	}
}

func WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade = autoMinorVersionUpgrade
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsAutoMinorVersionUpgrade", obj))
		},
	}
}

func WithKcpAwsAuthEnabled(authEnabled bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.AuthEnabled = authEnabled
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsAuthEnabled", obj))
		},
	}
}

func WithKcpAwsTransitEncryptionEnabled(transitEncryptionEnabled bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.TransitEncryptionEnabled = transitEncryptionEnabled
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsTransitEncryptionEnabled", obj))
		},
	}
}

func WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow *string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow = preferredMaintenanceWindow
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsPreferredMaintenanceWindow", obj))
		},
	}
}

func WithKcpAwsParameters(parameters map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				gcpRedisInstance.Spec.Instance.Aws.Parameters = parameters
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAwsParameters", obj))
		},
	}
}

func WithRedisInstanceAzure() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.Instance.Azure = &cloudcontrolv1beta1.RedisInstanceAzure{}
			}
		},
	}
}
func WithKcpAzureRedisVersion(redisVersion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				azureRedisInstance.Spec.Instance.Azure.RedisVersion = redisVersion
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureRedisVersion", obj))
		},
	}
}
func WithSKU(capacity int) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if azureRedisInstance, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				azureRedisInstance.Spec.Instance.Azure.SKU = cloudcontrolv1beta1.AzureRedisSKU{Capacity: capacity}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureRedisVersion", obj))
		},
	}
}
