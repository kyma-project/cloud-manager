package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareExpandShrink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.share.Size == state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	var err error
	if state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb < state.share.Size {
		logger.Info("Shrinking CCEE NfsInstance")
		state.ObjAsNfsInstance().Status.State = "Shrinking"
		err = state.cceeClient.ShareShrink(ctx, state.share.ID, state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb)
	} else {
		err = state.cceeClient.ShareExtend(ctx, state.share.ID, state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb)
		state.ObjAsNfsInstance().Status.State = "Extending"
	}

	if err != nil {
		logger.Error(err, "Failed to change size of CCEE NfsInstance")

		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Failed changing share size",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status after change size failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	logger.Info("Change size of CCEE NfsInstance success - requeuing")

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching CCEE NfsInstance status after change size success").
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
