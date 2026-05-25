package managedredis

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// setReconciling sets Ready=Unknown/Reconciling when a new generation is observed.
// Uses SuccessErrorNil so the action chain continues in the same reconcile cycle
// to perform the actual create/update work. ObservedGeneration is set later by
// updateStatus when the resource reaches Ready, providing proper staleness detection.
func setReconciling(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if obj.Generation == obj.Status.ObservedGeneration {
		return nil, ctx
	}

	return composed.UpdateStatus(obj).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionUnknown,
			Reason:  cloudcontrolv1beta1.ReasonProcessing,
			Message: "Reconciling",
		}).
		ErrorLogMessage("Error updating AzureManagedRedis status to Reconciling").
		SuccessErrorNil().
		Run(ctx, st)
}
