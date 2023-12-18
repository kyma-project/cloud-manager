package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/skr/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources/components/skr/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources/components/skr/pkg/common/genericActions"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func whenOutcomeError(ctx context.Context, state composed.State) error {
	obj := state.Obj().(genericActions.ObjWithStatus)
	objWithOutcome := state.Obj().(genericActions.ObjWithOutcome)
	outcome := objWithOutcome.GetOutcome()

	if outcome == nil || outcome.Type != cloudresourcesv1beta1.OutcomeTypeError {
		return nil
	}

	obj.SetStatusState(cloudresourcesv1beta1.ErrorState)
	meta.RemoveStatusCondition(obj.GetConditions(), string(cloudresourcesv1beta1.ConditionTypeReady))
	meta.RemoveStatusCondition(obj.GetConditions(), string(cloudresourcesv1beta1.ConditionTypeProcessing))
	meta.SetStatusCondition(obj.GetConditions(), metav1.Condition{
		Type:    string(cloudresourcesv1beta1.ConditionTypeError),
		Status:  metav1.ConditionTrue,
		Reason:  string(cloudresourcesv1beta1.ConditionReasonError),
		Message: outcome.Message,
	})

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return state.RequeueIfError(err)
	}

	return state.Stop(nil) // !!!important
}
