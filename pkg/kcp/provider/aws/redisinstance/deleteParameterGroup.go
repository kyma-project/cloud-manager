package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteParameterGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.parameterGroup == nil {
		return nil, nil
	}

	logger.
		WithValues("parameterGroupName", ptr.Deref(state.parameterGroup.CacheParameterGroupName, "")).
		Info("Deleting parameter group")

	err := state.awsClient.DeleteElastiCacheParameterGroup(ctx, ptr.Deref(state.parameterGroup.CacheParameterGroupName, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting parameter group", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
