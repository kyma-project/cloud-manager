package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteParameterGroup(getParamGroup func(*State) *elasticachetypes.CacheParameterGroup) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state := st.(*State)
		logger := composed.LoggerFromCtx(ctx)

		parameterGroup := getParamGroup(state)
		if parameterGroup == nil {
			return nil, ctx
		}

		logger.
			WithValues("parameterGroupName", ptr.Deref(parameterGroup.CacheParameterGroupName, "")).
			Info("Deleting parameter group")

		err := state.awsClient.DeleteElastiCacheParameterGroup(ctx, ptr.Deref(parameterGroup.CacheParameterGroupName, ""))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error deleting parameter group", ctx)
		}

		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}
}

func deleteMainParameterGroup() composed.Action {
	return deleteParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
	)
}

func deleteTempParameterGroup() composed.Action {
	return deleteParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
	)
}
