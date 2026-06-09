package rediscluster

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		redisInstance := state.ObjAsRedisCluster()
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
			logger.Error(err, "Error modifying cache parameters")
			meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: "Failed to modify cache parameters",
			})
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			updateErr := state.UpdateObjStatus(ctx)
			if updateErr != nil {
				return composed.LogErrorAndReturn(updateErr,
					"Error updating RedisCluster status due failed cache parameters modification",
					composed.StopWithRequeueDelay(util.Timing.T10000ms()),
					ctx,
				)
			}

			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
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
