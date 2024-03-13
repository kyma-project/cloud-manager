package nfsinstance

import (
	"context"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func waitEfsDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.efs == nil {
		logger.Info("The EFS not found")
		return nil, nil
	}
	if state.efs.LifeCycleState == efsTypes.LifeCycleStateDeleted {
		logger.Info("The EFS is the deleted state")
		return nil, nil
	}

	if state.efs.LifeCycleState == efsTypes.LifeCycleStateError {
		return composed.UpdateStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonUnknown,
				Message: "EFS in error state",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, st)
	}

	logger.
		WithValues("lifeCycleState", state.efs.LifeCycleState).
		Info("Waiting EFS gets deleted")

	return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
}
