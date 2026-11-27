package cloudresources

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
		MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
			cr.Status.State = cloudresourcesv1beta1.ModuleStateReady
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:               cloudresourcesv1beta1.ConditionTypeReady,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: cr.Generation,
				Reason:             cloudresourcesv1beta1.ConditionTypeReady,
				Message:            cloudresourcesv1beta1.ConditionTypeReady,
			})
		}).
		OnFailure(
			composed.Log("Error patching CloudResources with Ready condition true"),
			composed.Requeue,
		).
		OnSuccess(composed.Forget).
		OnStatusChanged(composed.Log("CloudResources patched to Ready true")).
		Run(ctx, state.Cluster().K8sClient())
}
