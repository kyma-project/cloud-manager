package awsrediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitKcpRedisClusterDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	awsRedisCluster := state.ObjAsAwsRedisCluster()

	if state.KcpRedisCluster == nil {
		logger.Info("Kcp RedisCluster is deleted")
		return nil, nil
	}

	kcpCondErr := meta.FindStatusCondition(state.KcpRedisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if kcpCondErr != nil {
		awsRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(awsRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AwsRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR AwsRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.Info("Waiting for Kcp RedisCluster to be deleted")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
