package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func whenBeingDeleted(ctx context.Context, state composed.State) error {
	instanceBeingDeleted := !state.Obj().GetDeletionTimestamp().IsZero()
	if !instanceBeingDeleted {
		return nil
	}

	objWithOutcome := state.Obj().(genericActions.ObjWithOutcome)
	obj := state.Obj().(genericActions.ObjWithStatus)
	outcome := objWithOutcome.GetOutcome()

	if outcome == nil {
		return state.Stop(nil) // waiting for the outcome, setting it will trigger a new loop
	}

	hasDeletedOutcome := outcome != nil && outcome.Type == cloudresourcesv1beta1.OutcomeTypeDeleted
	if !hasDeletedOutcome {
		return state.Stop(nil) // waiting for the outcome, setting it will trigger a new loop
	}

	obj.SetStatusState(cloudresourcesv1beta1.UnknownState)

	meta.RemoveStatusCondition(obj.GetConditions(), string(cloudresourcesv1beta1.ConditionTypeReady))
	meta.RemoveStatusCondition(obj.GetConditions(), string(cloudresourcesv1beta1.ConditionTypeError))
	meta.SetStatusCondition(obj.GetConditions(), metav1.Condition{
		Type:    string(cloudresourcesv1beta1.ConditionTypeDeleted),
		Status:  metav1.ConditionTrue,
		Reason:  string(cloudresourcesv1beta1.ConditionReasonDeleted),
		Message: "Resource is deleted",
	})

	err := state.UpdateObjStatus(ctx)
	if client.IgnoreNotFound(err) != nil {
		return state.RequeueIfError(err)
	}

	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	err = state.UpdateObj(ctx)
	if client.IgnoreNotFound(err) != nil {
		return state.RequeueIfError(err)
	}

	return state.Stop(nil) // !!!important
}
