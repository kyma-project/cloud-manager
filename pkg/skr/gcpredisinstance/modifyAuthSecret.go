package gcpredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		logger.Info("cant modify auth secret, not found")
		return nil, nil
	}

	currentSecretData := state.AuthSecret.Data
	desiredSecretData := getAuthSecretData(state.KcpRedisInstance)

	if areByteMapsEqual(currentSecretData, desiredSecretData) {
		return nil, nil
	}

	state.AuthSecret.Data = desiredSecretData

	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating secret for GcpRedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for GcpRedisInstance updated")

	return nil, nil
}
