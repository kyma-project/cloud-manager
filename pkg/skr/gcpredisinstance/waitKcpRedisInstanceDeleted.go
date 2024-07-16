package gcpredisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitKcpRedisInstanceDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	gcpRedisInstance := state.ObjAsGcpRedisInstance()

	if state.KcpRedisInstance == nil {
		logger.Info("Kcp NfsInstance is deleted")
		return nil, nil
	}

	kcpCondErr := meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	if kcpCondErr != nil {
		gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(gcpRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating GcpRedisInstance status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR GcpRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.Info("Waiting for Kcp NfsInstance to be deleted")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
