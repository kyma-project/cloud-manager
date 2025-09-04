package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func modifyParameterGroup(
	getParamGroup func(*State) *elasticachetypes.CacheParameterGroup,
	getParamGroupCurrentParams func(*State) []elasticachetypes.Parameter,
	getParamGroupDefaultParams func(*State) []elasticachetypes.Parameter,
	name string,
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state := st.(*State)
		logger := composed.LoggerFromCtx(ctx)

		redisInstance := state.ObjAsRedisInstance()
		parameterGroup := getParamGroup(state)

		if parameterGroup == nil {
			return nil, ctx
		}
		currentParameters := getParamGroupCurrentParams(state)
		defaultParameters := getParamGroupDefaultParams(state)

		currentParametersMap := MapParameters(currentParameters)
		defaultParametersMap := MapParameters(defaultParameters)

		desiredParametersMap := GetDesiredParameters(defaultParametersMap, redisInstance.Spec.Instance.Aws.Parameters)
		forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

		if len(forUpdateParameters) == 0 {
			return nil, ctx
		}

		logger.Info("Modifying cache parameters")
		err := state.awsClient.ModifyElastiCacheParameterGroup(ctx, name, ToParametersSlice(forUpdateParameters))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error modifying cache parameters", ctx)
		}

		return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
	}
}

func modifyMainParameterGroup(state *State) composed.Action {
	return modifyParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State) []elasticachetypes.Parameter { return s.parameterGroupCurrentParams },
		func(s *State) []elasticachetypes.Parameter { return s.parameterGroupFamilyDefaultParams },
		GetAwsElastiCacheParameterGroupName(state.Obj().GetName()),
	)
}

func modifyTempParameterGroup(state *State) composed.Action {
	return modifyParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State) []elasticachetypes.Parameter { return s.tempParameterGroupCurrentParams },
		func(s *State) []elasticachetypes.Parameter { return s.tempParameterGroupFamilyDefaultParams },
		GetAwsElastiCacheTempParameterGroupName(state.Obj().GetName()),
	)
}
