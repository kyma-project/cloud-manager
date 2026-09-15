package alicloudrediscluster

import (
	"bytes"
	"context"
	"maps"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func modifyAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		// The secret was just created but the cache hasn't caught up yet. Requeue
		// so modifyAuthSecret runs again with a fresh load rather than skipping and
		// letting updateStatus mark the SKR Ready with an incomplete secret.
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	currentSecretData := state.AuthSecret.Data
	desiredSecretData := state.GetAuthSecretData()

	desiredLabels := getAuthSecretLabels(state.ObjAsAlicloudRedisCluster())
	desiredAnnotations := getAuthSecretAnnotations(state.ObjAsAlicloudRedisCluster())

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
		return composed.LogErrorAndReturn(err, "Error updating secret for AlicloudRedisCluster", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for AlicloudRedisCluster updated")

	return nil, ctx
}
