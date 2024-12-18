package nfsinstance

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareWaitAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.share.Status == "available" {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	if state.share.Status != "error" {
		logger.Info("Waiting for CCEE share to be available")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	// share is in error state
	logger.Error(errors.New("ccee share in error state"), "CCEE share error")

	state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: "CCEE Share in error state",
		}).
		ErrorLogMessage("Error parching CCEE NfsInstance status after share in error state").
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
		Run(ctx, state)
}
