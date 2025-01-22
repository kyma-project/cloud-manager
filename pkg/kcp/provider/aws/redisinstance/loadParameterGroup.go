package redisinstance

import (
	"context"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadParameterGroup(
	getParamGroup func(*State) *elasticacheTypes.CacheParameterGroup,
	setParamGroup func(*State, *elasticacheTypes.CacheParameterGroup),
	name string,
) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		paramGroup := getParamGroup(state)
		if paramGroup != nil {
			return nil, nil
		}

		logger := composed.LoggerFromCtx(ctx)

		list, err := state.awsClient.DescribeElastiCacheParameterGroup(ctx, name)
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error listing parameter groups", ctx)
		}

		if len(list) > 0 {
			foundParamGroup := &list[0]
			setParamGroup(state, foundParamGroup)
			logger = logger.WithValues("parameterGroupName", ptr.Deref(foundParamGroup.CacheParameterGroupName, ""))
			logger.Info("ElastiCache parameter group found and loaded")
			return nil, composed.LoggerIntoCtx(ctx, logger)
		}

		logger.Info("ElastiCache parameter group not found")

		return nil, nil
	}
}

func loadMainParameterGroup(state *State) composed.Action {
	return loadParameterGroup(
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.parameterGroup },
		func(s *State, paramGroup *elasticacheTypes.CacheParameterGroup) { s.parameterGroup = paramGroup },
		GetAwsElastiCacheParameterGroupName(state.Obj().GetName()),
	)
}

func loadTempParameterGroup(state *State) composed.Action {
	return loadParameterGroup(
		func(s *State) *elasticacheTypes.CacheParameterGroup { return s.tempParameterGroup },
		func(s *State, paramGroup *elasticacheTypes.CacheParameterGroup) { s.tempParameterGroup = paramGroup },
		GetAwsElastiCacheTempParameterGroupName(state.Obj().GetName()),
	)
}
