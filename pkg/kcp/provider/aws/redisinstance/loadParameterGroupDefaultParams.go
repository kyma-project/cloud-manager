package redisinstance

import (
	"context"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadParameterGroupDefaultParams(
	getParamGroup func(*State) *elasticacheTypes.CacheParameterGroup,
	setDefaultParams func(*State, []elasticacheTypes.Parameter),
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		paramGroup := getParamGroup(state)
		if paramGroup == nil {
			return nil, nil
		}

		defaultParams, err := state.awsClient.DescribeEngineDefaultParameters(ctx, ptr.Deref(paramGroup.CacheParameterGroupFamily, ""))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error getting default parameters", ctx)
		}

		setDefaultParams(state, defaultParams)

		return nil, nil
	}
}

func loadMainParameterGroupDefaultParams() composed.Action {
	return loadParameterGroupDefaultParams(
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State, parameters []elasticacheTypes.Parameter) { s.parameterGroupDefaultParams = parameters },
	)
}

func loadTempParameterGroupDefaultParams() composed.Action {
	return loadParameterGroupDefaultParams(
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State, parameters []elasticacheTypes.Parameter) {
			s.tempParameterGroupDefaultParams = parameters
		},
	)
}
