package subnet

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitNetworkReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// it is in general flow in reconciler, so we need this
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.network.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.network.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError) {
		logger.Info("KCP GcpSubnet kyma network has error condition")
		state.network.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(state.network).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Network in error state",
			}).
			ErrorLogMessage("Error updating KCP GcpSubnet status after setting kyma network in error state status").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	logger.Info("Waiting KCP GcpSubnet kyma network ready state")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
