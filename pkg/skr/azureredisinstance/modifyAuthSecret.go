package azureredisinstance

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

	desiredLabels := getAuthSecretLabels(state.ObjAsAzureRedisInstance())
	desiredAnnotations := getAuthSecretAnnotations(state.ObjAsAzureRedisInstance())

	dataChanged := !maps.EqualFunc(currentSecretData, desiredSecretData, func(l, r []byte) bool { return bytes.Equal(l, r) })
	labelsChanged := !maps.Equal(state.AuthSecret.Labels, desiredLabels)
	annotationsChanged := !maps.Equal(state.AuthSecret.Annotations, desiredAnnotations)

	if !dataChanged && !labelsChanged && !annotationsChanged {
		return nil, ctx
	}

	state.AuthSecret.Data = desiredSecretData
	state.AuthSecret.Labels = desiredLabels
	state.AuthSecret.Annotations = desiredAnnotations

	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating secret for AzureRedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for AzureRedisInstance updated")

	return nil, ctx
}
