package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadParameterGroupCurrentParams(
	getParamGroup func(*State) *elasticachetypes.CacheParameterGroup,
	setCurrentParams func(*State, []elasticachetypes.Parameter),
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		paramGroup := getParamGroup(state)
		if paramGroup == nil {
			return nil, nil
		}

		currentParams, err := state.awsClient.DescribeElastiCacheParameters(ctx, ptr.Deref(paramGroup.CacheParameterGroupName, ""))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error getting current parameters", ctx)
		}

		setCurrentParams(state, currentParams)

		return nil, nil
	}
}

func loadMainParameterGroupCurrentParams() composed.Action {
	return loadParameterGroupCurrentParams(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State, parameters []elasticachetypes.Parameter) { s.parameterGroupCurrentParams = parameters },
	)
}

func loadTempParameterGroupCurrentParams() composed.Action {
	return loadParameterGroupCurrentParams(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State, parameters []elasticachetypes.Parameter) {
			s.tempParameterGroupCurrentParams = parameters
		},
	)
}
