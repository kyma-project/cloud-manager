package reconcile

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/api/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func whenNoFinalizer(ctx context.Context, state composed.State) error {
	hasFinalizer := controllerutil.ContainsFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)

	if hasFinalizer {
		return nil
	}

	obj := state.Obj().(genericActions.ObjWithStatus)

	obj.SetStatusState(cloudresourcesv1beta1.UnknownState)

	meta.SetStatusCondition(obj.GetConditions(), metav1.Condition{
		Type:    string(cloudresourcesv1beta1.ConditionTypeProcessing),
		Status:  metav1.ConditionTrue,
		Reason:  string(cloudresourcesv1beta1.ConditionReasonProcessing),
		Message: "Resource is being provisioned",
	})

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return state.RequeueIfError(err)
	}

	controllerutil.AddFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)

	err = state.UpdateObj(ctx)
	if err != nil {
		return state.RequeueIfError(err)
	}

	return state.StopWithRequeue() // !!!important
}
