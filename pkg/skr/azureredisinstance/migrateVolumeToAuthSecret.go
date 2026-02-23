package azureredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// migrateVolumeToAuthSecret handles backward compatibility for the Volume field.
func migrateVolumeToAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAzureRedisInstance()

	if state.AuthSecretDetails != nil {
		return nil, ctx
	}

	state.AuthSecretDetails = obj.Spec.Volume

	if obj.Spec.AuthSecret != nil {
		state.AuthSecretDetails = obj.Spec.AuthSecret

		if obj.Spec.Volume != nil {
			logger.Info("Both 'volume' (deprecated) and 'authSecret' fields are set. Using 'authSecret' and ignoring 'volume'. Please remove 'spec.volume' from your Git repository.")
		}
	} else if obj.Spec.Volume != nil {
		logger.Info("DEPRECATION WARNING: 'spec.volume' is deprecated. Please update your Git repository to use 'spec.authSecret' instead. Functionality will continue to work with the old field name.")
	}

	return nil, ctx
}
