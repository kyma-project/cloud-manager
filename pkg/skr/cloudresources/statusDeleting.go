package cloudresources

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	return composed.NewStatusPatcherComposed(state.ObjAsCloudResources()).
		MutateStatus(func(cr *cloudresourcesv1beta1.CloudResources) {
			cr.Status.State = cloudresourcesv1beta1.ModuleStateDeleting
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:               cloudresourcesv1beta1.ConditionTypeReady,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: cr.Generation,
				Reason:             cloudresourcesv1beta1.ReasonDeleting,
				Message:            cloudresourcesv1beta1.ReasonDeleting,
			})
		}).
		OnFailure(
			composed.Log("Error patching CloudResources status.state to Deleting"),
			composed.Requeue,
		).
		OnSuccess(composed.Continue).
		Run(ctx, state.Cluster().K8sClient())
}
