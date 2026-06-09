package redisinstance

import (
	"context"

	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/ptr"
)

func createAuthTokenSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.authTokenValue != nil {
		return nil, ctx
	}

	redisInstance := state.ObjAsRedisInstance()

	if !redisInstance.Spec.Instance.Aws.AuthEnabled {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)

	err := state.awsClient.CreateAuthTokenSecret(ctx, GetAwsAuthTokenSecretName(state.Obj().GetName()), []secretsmanagertypes.Tag{
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
		logger.Error(err, "Error creating authToken secret")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to create authToken secret",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		updateErr := state.UpdateObjStatus(ctx)
		if updateErr != nil {
			return composed.LogErrorAndReturn(updateErr,
				"Error updating RedisInstance status due failed authToken secret creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger.Info("AuthToken secret created")

	return composed.StopWithRequeue, nil
}
