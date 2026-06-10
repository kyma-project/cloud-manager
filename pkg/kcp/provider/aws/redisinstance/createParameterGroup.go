package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/ptr"
)

func createParameterGroup(getParamGroup func(*State) *elasticachetypes.CacheParameterGroup, name string) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)

		if getParamGroup(state) != nil {
			return nil, ctx
		}

		redisInstance := state.ObjAsRedisInstance()

		logger := composed.LoggerFromCtx(ctx)

		family := GetAwsElastiCacheParameterGroupFamily(redisInstance.Spec.Instance.Aws.EngineVersion)

		out, err := state.awsClient.CreateElastiCacheParameterGroup(ctx, name, family, []elasticachetypes.Tag{
			{
				Key:   ptr.To(common.TagCloudManagerRemoteName),
				Value: new(redisInstance.Spec.RemoteRef.String()),
			},
			{
				Key:   ptr.To(common.TagCloudManagerName),
				Value: new(state.Name().String()),
			},
			{
				Key:   ptr.To(common.TagScope),
				Value: new(redisInstance.Spec.Scope.Name),
			},
			{
				Key:   ptr.To(common.TagShoot),
				Value: new(state.Scope().Spec.ShootName),
			},
		})
		if err != nil {
			logger.Error(err, "Error creating parameter group")
			meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: "Failed to create parameter group",
			})
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			updateErr := state.UpdateObjStatus(ctx)
			if updateErr != nil {
				return composed.LogErrorAndReturn(updateErr,
					"Error updating RedisInstance status due failed parameter group creation",
					composed.StopWithRequeueDelay(util.Timing.T10000ms()),
					ctx,
				)
			}

			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}

		logger = logger.WithValues("parameterGroupName", out.CacheParameterGroup.CacheParameterGroupName)
		logger.Info("Parameter group created")

		return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
	}
}

func createMainParameterGroup(state *State) composed.Action {
	return createParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.parameterGroup },
		GetAwsElastiCacheParameterGroupName(state.Obj().GetName()),
	)
}

func createTempParameterGroup(state *State) composed.Action {
	return createParameterGroup(
		func(s *State) *elasticachetypes.CacheParameterGroup { return s.tempParameterGroup },
		GetAwsElastiCacheTempParameterGroupName(state.Obj().GetName()),
	)
}
