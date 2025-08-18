package awsrediscluster

import (
	"bytes"
	"context"
	"maps"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		logger.Info("cant modify auth secret, not found")
		return nil, ctx
	}

	currentSecretData := state.AuthSecret.Data
	desiredSecretData := state.GetAuthSecretData()

	if maps.EqualFunc(currentSecretData, desiredSecretData, func(l, r []byte) bool { return bytes.Equal(l, r) }) {
		return nil, ctx
	}

	state.AuthSecret.Data = desiredSecretData

	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating secret for AwsRedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for AwsRedisCluster updated")

	return nil, ctx
}
