package reconcile

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	"github.com/kyma-project/cloud-resources-manager/pkg/common/genericActions"
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

		return fmt.Errorf("only one instance of CloudResources is allowed (current served instance: %s",
			served.Name)
	}

	if obj.Status.Served == cloudresourcesv1beta1.ServedFalse {
		return state.Stop(nil)
	}

	return nil
}
