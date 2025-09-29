package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kcpNetworkWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.network.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) {
		logger.Info("KCP IpRange network is Ready")
		return nil, ctx
	}

	if meta.IsStatusConditionTrue(state.network.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError) {
		logger.Info("KCP IpRange network has error condition")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Network in error state",
			}).
			ErrorLogMessage("Error patching KCP IpRange status after setting network in error state status").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	logger.Info("Waiting KCP Network ready state for IpRange")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
