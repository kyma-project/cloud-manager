package azurerediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// migrateVolumeToAuthSecret handles backward compatibility for the Volume field.
func migrateVolumeToAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureRedisCluster()

	if state.AuthSecretDetails != nil {
		return nil, ctx
	}

	state.AuthSecretDetails = obj.Spec.Volume

	if obj.Spec.AuthSecret != nil {
		state.AuthSecretDetails = obj.Spec.AuthSecret
	}

	return nil, ctx
}
