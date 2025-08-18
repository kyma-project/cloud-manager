package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadParameterGroupFamilyDefaultParams(
	getParamGroup func(*State) *elasticachetypes.CacheParameterGroup,
	setDefaultParams func(*State, []elasticachetypes.Parameter),
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		paramGroup := getParamGroup(state)
		if paramGroup == nil {
			return nil, ctx
		}

		defaultParams, err := state.awsClient.DescribeEngineDefaultParameters(ctx, ptr.Deref(paramGroup.CacheParameterGroupFamily, ""))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error getting default parameters", ctx)
		}

		setDefaultParams(state, defaultParams)

		return nil, ctx
	}
}

func loadMainParameterGroupFamilyDefaultParams() composed.Action {
	return loadParameterGroupFamilyDefaultParams(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State, parameters []elasticachetypes.Parameter) {
			s.parameterGroupFamilyDefaultParams = parameters
		},
	)
}

func loadTempParameterGroupFamilyDefaultParams() composed.Action {
	return loadParameterGroupFamilyDefaultParams(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State, parameters []elasticachetypes.Parameter) {
			s.tempParameterGroupFamilyDefaultParams = parameters
		},
	)
}
