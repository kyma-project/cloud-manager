package azureredisinstance

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func getEffectiveAuthSecret(obj *cloudresourcesv1beta1.AzureRedisInstance) *cloudresourcesv1beta1.RedisAuthSecretSpec {
	if obj.Spec.AuthSecret != nil {
		return obj.Spec.AuthSecret
	}
	return obj.Spec.Volume
}

// migrateVolumeToAuthSecret handles backward compatibility for the Volume field.
func migrateVolumeToAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAzureRedisInstance()

	// If both fields are set, prefer AuthSecret but don't modify the resource
	if obj.Spec.Volume != nil && obj.Spec.AuthSecret != nil {
		logger.Info("Both 'volume' (deprecated) and 'authSecret' fields are set. Using 'authSecret' and ignoring 'volume'. Please remove 'spec.volume' from your Git repository.")
		return nil, ctx
	}

	// If only Volume is set, use it but don't modify the resource (ArgoCD-safe)
	if obj.Spec.Volume != nil && obj.Spec.AuthSecret == nil {
		logger.Info("DEPRECATION WARNING: 'spec.volume' is deprecated. Please update your Git repository to use 'spec.authSecret' instead. Functionality will continue to work with the old field name.")
		return nil, ctx
	}

	return nil, ctx
}
