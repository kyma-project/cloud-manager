package redisinstance

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createUserGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.userGroup != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisInstance()

	out, err := state.awsClient.CreateUserGroup(ctx, GetAwsElastiCacheParameterGroupName(state.Obj().GetName()), []types.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: new(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: new(redisInstance.Spec.RemoteRef.String()),
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
		if awsmeta.IsErrorRetryable(err) {
			return awsmeta.LogErrorAndReturn(err, "Error creating user group", ctx)
		}
		logger.Error(err, "Error creating user group")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create user group",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		updateErr := state.UpdateObjStatus(ctx)
		if updateErr != nil {
			return composed.LogErrorAndReturn(updateErr,
				"Error updating RedisInstance status due failed user group creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger = logger.WithValues("userGroupId", out.UserGroupId)
	logger.Info("User group created")

	return composed.StopWithRequeue, nil
}
