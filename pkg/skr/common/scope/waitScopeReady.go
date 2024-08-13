package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func waitScopeReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	hasReady := meta.FindStatusCondition(state.Scope().Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if hasReady != nil {
		// scope is ready
		// * set status.State back to empty
		// * remove ConditionTypeWaitScopeReady
		// * continue (return nil error)

		if state.ObjWithConditionsAndState().State() != cloudresourcesv1beta1.StateWaitingScopeReady {
			return nil, nil
		}

		state.ObjWithConditionsAndState().SetState("")
		return composed.UpdateStatus(state.ObjWithConditionsAndState()).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeWaitScopeReady).
			SuccessError(composed.StopWithRequeue).
			ErrorLogMessage("Error updating object status after removing ").
			Run(ctx, state)
	}

	// scope is not ready
	// * set status.State to WaitingScopeReady
	// * set condition WaitScopeReady
	// * requeue and delay

	state.ObjWithConditionsAndState().SetState(cloudresourcesv1beta1.StateWaitingScopeReady)

	return composed.UpdateStatus(state.ObjWithConditionsAndState()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWaitScopeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeWaitScopeReady,
			Message: "Wait Scope to be Ready",
		}).
		SuccessError(composed.StopWithRequeueDelay(200*time.Millisecond)).
		ErrorLogMessage("Error updating object status after removing ").
		Run(ctx, state)
}
