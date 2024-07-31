package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func modifyParameterGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	if state.parameterGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentParameters, err := state.awsClient.DescribeElastiCacheParameters(ctx, GetAwsElastiCacheParameterGroupName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error getting current parameters", ctx)
	}

	family := GetAwsElastiCacheParameterGroupFamily(redisInstance.Spec.Instance.Aws.EngineVersion)
	defaultParameters, err := state.awsClient.DescribeEngineDefaultParameters(ctx, family)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error getting default parameters", ctx)
	}

	currentParametersMap := MapParameters(currentParameters)
	defaultParametersMap := MapParameters(defaultParameters)

	desiredParametersMap := GetDesiredParameters(defaultParametersMap, redisInstance.Spec.Instance.Aws.Parameters)
	forUpdateParameters := GetMissmatchedParameters(currentParametersMap, desiredParametersMap)

	if len(forUpdateParameters) > 0 {
		logger.Info("Modifying cache parameters")
		err = state.awsClient.ModifyElastiCacheParameterGroup(ctx, GetAwsElastiCacheParameterGroupName(state.Obj().GetName()), ToParametersSlice(forUpdateParameters))
	}
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error modifying cache parameters", ctx)
	}

	return nil, nil
}
