package network

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// logLogicalErrorOnManagedNetworkMissing is a logical safeguard that is protecting the flow. All network references
// must be reconciled before this action and stopped. If a network reference reach this action it is considered a
// logical exception and a development flow
func logLogicalErrorOnManagedNetworkMissing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*state)
	logger := composed.LoggerFromCtx(ctx)

	if state.Scope() == nil || state.Scope().Name == "" {
		err := errors.New("scope not found")
		logger.Error(err, "Logical error")
		state.ObjAsNetwork().Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.PatchStatus(state.ObjAsNetwork()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
				Message: "Scope not found",
			}).
			ErrorLogMessage("Error patching KCP Network status with scope not found condition").
			SuccessError(composed.StopAndForget).
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	if state.ObjAsNetwork().Spec.Network.Managed == nil {
		err := errors.New("expected managed network, but none is present in state")
		logger.Error(err, "Logical error")
		return composed.StopAndForget, nil
	}

	if state.ObjAsNetwork().Spec.Network.Reference != nil {
		err := errors.New("did not expect network reference, but it is present in state")
		logger.Error(err, "Logical error")
		return composed.StopAndForget, nil
	}

	return nil, nil
}
