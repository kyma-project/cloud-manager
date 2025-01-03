package awsredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	util "github.com/kyma-project/cloud-manager/pkg/util"
)

func modifyAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		logger.Info("cant modify auth secret, not found")
		return nil, nil
	}

	currentSecretData := state.AuthSecret.Data
	desiredSecretData := state.GetAuthSecretData()

	if util.AreByteMapsEqual(currentSecretData, desiredSecretData) {
		return nil, nil
	}

	state.AuthSecret.Data = desiredSecretData

	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating secret for AwsRedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for AwsRedisInstance updated")

	return nil, nil
}
