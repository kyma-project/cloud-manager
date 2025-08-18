package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func modifyAuthEnabled(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisCluster()
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentAuthEnabled := ptr.Deref(state.elastiCacheReplicationGroup.AuthTokenEnabled, false)
	desiredAuthEnabled := redisInstance.Spec.Instance.Aws.AuthEnabled

	if currentAuthEnabled == desiredAuthEnabled && len(state.elastiCacheReplicationGroup.UserGroupIds) == 0 {
		return nil, ctx
	}

	if desiredAuthEnabled && state.authTokenValue == nil {
		return composed.StopWithRequeue, nil
	}

	if !desiredAuthEnabled && state.authTokenValue != nil {
		logger.Info("Deleting authToken secret")
		err := state.awsClient.DeleteAuthTokenSecret(ctx, *state.authTokenValue.Name)
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error deleting authToken secret", ctx)
		}
	}

	state.UpdateAuthEnabled(desiredAuthEnabled)

	return nil, ctx
}
