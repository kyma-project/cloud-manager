package cloudresources

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func handleServed(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list := &cloudresourcesv1beta1.CloudResourcesList{}
	err := state.Cluster().ApiReader().List(ctx, list)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing CloudResources", composed.StopWithRequeue, ctx)
	}

	var served *cloudresourcesv1beta1.CloudResources
	for _, item := range list.Items {
		if item.Status.Served == cloudresourcesv1beta1.ServedTrue {
			served = &item
			break
		}
	}

	if served != nil && state.Obj().GetName() == served.Name && state.Obj().GetNamespace() == served.Namespace {
		// nothing to do, the obj is the served one
		return nil, ctx
	}

	if state.ObjAsCloudResources().Status.Served == cloudresourcesv1beta1.ServedFalse {
		// we're not handling not served objects
		return composed.StopAndForget, ctx
	}

	if served == nil {
		// none is served so far, this obj will be the one
		return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
			MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
				cr.Status.Served = cloudresourcesv1beta1.ServedTrue
			}).
			OnStatusChanged(composed.Log("Set CloudResources Served to True")).
			OnFailure(
				composed.Log("Error patching CloudResources status.served to true"),
				composed.Requeue,
			).
			OnSuccess(composed.Continue).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
		MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
			cr.Status.Served = cloudresourcesv1beta1.ServedFalse
			cr.Status.State = cloudresourcesv1beta1.ModuleStateWarning
			cr.Status.Conditions = nil
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:               cloudresourcesv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: cr.Generation,
				Reason:             cloudresourcesv1beta1.ReasonOtherIsServed,
				Message:            fmt.Sprintf("only one instance of CloudResources is allowed (current served instance: %s", served.Name),
			})
		}).
		OnFailure(
			composed.Log("Error patching CloudResources status.served to false"),
			composed.Requeue,
		).
		OnSuccess(composed.Forget).
		Run(ctx, state.Cluster().K8sClient())
}
