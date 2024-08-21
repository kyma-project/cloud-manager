package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func networkStopWhenNotFound(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.network != nil {
		return nil, nil
	}

	networkId, _ := state.ObjAsNfsInstance().GetStateData(StateDataNetworkId)
	logger.
		WithValues(
			"networkName", state.Scope().Spec.Scope.OpenStack.VpcNetwork,
			"networkId", networkId,
		).
		Info("CCEE network not found")

	state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SetCondition(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: "Network not found",
		}).
		ErrorLogMessage("Error patching CCEE NfsInstance with network not found condition").
		Run(ctx, state)
}
