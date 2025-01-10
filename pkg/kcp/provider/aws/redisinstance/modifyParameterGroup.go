package redisinstance

import (
	"context"

	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func modifyParameterGroup(getParamGroup func(*State) *elasticacheTypes.CacheParameterGroup, name string) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state := st.(*State)
		logger := composed.LoggerFromCtx(ctx)

		redisInstance := state.ObjAsRedisInstance()
		parameterGroup := getParamGroup(state)

		if parameterGroup == nil {
			return nil, nil
		}
		currentParameters, err := state.awsClient.DescribeElastiCacheParameters(ctx, name)
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error getting current parameters", ctx)
		}

		family := ptr.Deref(parameterGroup.CacheParameterGroupFamily, "")
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
			err = state.awsClient.ModifyElastiCacheParameterGroup(ctx, name, ToParametersSlice(forUpdateParameters))
		}
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error modifying cache parameters", ctx)
		}

		return nil, nil
	}
}
