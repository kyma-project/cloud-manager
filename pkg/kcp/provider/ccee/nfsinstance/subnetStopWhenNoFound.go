package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func subnetStopWhenNoFound(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, nil
	}

	networkId, _ := state.ObjAsNfsInstance().GetStateData(StateDataNetworkId)
	subnetId, _ := state.ObjAsNfsInstance().GetStateData(StateDataSubnetId)
	logger.
		WithValues(
			"networkName", state.Scope().Spec.Scope.OpenStack.VpcNetwork,
			"networkId", networkId,
			"subnetId", subnetId,
		).
		Info("CCEE subnet not found")

	state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SetCondition(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: "Subnet not found",
		}).
		ErrorLogMessage("Error patching CCEE NfsInstance with network not found condition").
		Run(ctx, state)
}
