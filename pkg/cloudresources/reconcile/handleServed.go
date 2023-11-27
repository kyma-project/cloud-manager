package reconcile

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleServed(ctx context.Context, state composed.State) error {
	obj := state.Obj().(*cloudresourcesv1beta1.CloudResources)
	served := state.(genericActions.StateWithCloudResources).ServedCloudResources()
	if served != nil && obj.Name == served.Name {
		// the obj is the served one, nothing to do
		return nil
	}

	if obj.Status.Served == "" {
		if served == nil {
			obj.Status.Served = cloudresourcesv1beta1.ServedTrue
			err := state.UpdateObjStatus(ctx)
			return state.RequeueIfError(err, "error updating status.served to true for CloudResources %s", obj.Name)
		}

		msg := fmt.Sprintf("only one instance of CloudResources is allowed (current served instance: %s",
			served.Name)

		obj.SetStatusState(cloudresourcesv1beta1.ErrorState)

		meta.SetStatusCondition(obj.GetConditions(), metav1.Condition{
			Type:    string(cloudresourcesv1beta1.ConditionTypeError),
			Status:  metav1.ConditionTrue,
			Reason:  string(cloudresourcesv1beta1.ConditionReasonError),
			Message: msg,
		})

		err := state.UpdateObjStatus(ctx)

		return state.RequeueIfError(err)
	}

	if obj.Status.Served == cloudresourcesv1beta1.ServedFalse {
		return state.Stop(nil) // we're not reconciling objects that are not served
	}

	return nil
}
