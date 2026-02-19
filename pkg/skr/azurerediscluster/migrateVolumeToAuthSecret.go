package azurerediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// migrateVolumeToAuthSecret handles backward compatibility for the Volume field.
func migrateVolumeToAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAzureRedisCluster()

	if obj.Spec.Volume != nil && obj.Spec.AuthSecret != nil {
		logger.Info("Both 'volume' (deprecated) and 'authSecret' fields are set. Using 'authSecret' and ignoring 'volume'.")
		obj.Spec.Volume = nil
		err := state.UpdateObj(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating AzureRedisCluster after clearing deprecated 'volume' field", composed.StopWithRequeue, ctx)
		}
		return nil, ctx
	}

	// Migrate from Volume to AuthSecret if Volume is set and AuthSecret is not
	if obj.Spec.Volume != nil && obj.Spec.AuthSecret == nil {
		logger.Info("Migrating deprecated 'volume' field to 'authSecret'")
		obj.Spec.AuthSecret = obj.Spec.Volume
		obj.Spec.Volume = nil

		err := state.UpdateObj(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error migrating 'volume' to 'authSecret'", composed.StopWithRequeue, ctx)
		}

		logger.Info("Successfully migrated 'volume' to 'authSecret'. Please update your resource definition to use 'spec.authSecret' instead of 'spec.volume'.")
		return nil, ctx
	}

	return nil, ctx
}
