package rediscluster

import (
	"context"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func modifyParameterGroup(
	getParamGroup func(*State) *elasticacheTypes.CacheParameterGroup,
	getParamGroupCurrentParams func(*State) []elasticacheTypes.Parameter,
	getParamGroupDefaultParams func(*State) []elasticacheTypes.Parameter,
	name string,
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state := st.(*State)
		logger := composed.LoggerFromCtx(ctx)

		redisInstance := state.ObjAsRedisCluster()
		parameterGroup := getParamGroup(state)

		if parameterGroup == nil {
			return nil, nil
		}
		currentParameters := getParamGroupCurrentParams(state)
		defaultParameters := getParamGroupDefaultParams(state)

		currentParametersMap := MapParameters(currentParameters)
		defaultParametersMap := MapParameters(defaultParameters)

		desiredParametersMap := GetDesiredParameters(defaultParametersMap, redisInstance.Spec.Instance.Aws.Parameters)
		forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

		if len(forUpdateParameters) == 0 {
			return nil, nil
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
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State) []elasticacheTypes.Parameter { return s.parameterGroupCurrentParams },
		func(s *State) []elasticacheTypes.Parameter { return s.parameterGroupFamilyDefaultParams },
		GetAwsElastiCacheParameterGroupName(state.Obj().GetName()),
	)
}

func modifyTempParameterGroup(state *State) composed.Action {
	return modifyParameterGroup(
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State) []elasticacheTypes.Parameter { return s.tempParameterGroupCurrentParams },
		func(s *State) []elasticacheTypes.Parameter { return s.tempParameterGroupFamilyDefaultParams },
		GetAwsElastiCacheTempParameterGroupName(state.Obj().GetName()),
	)
}
