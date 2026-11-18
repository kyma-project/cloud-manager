package scope

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if state.ObjWithStatus() == nil {
		// unsupported type, likely older API style with object
		// implementing composed.ObjWithConditionsAndState
		return nil, ctx
	}

	sp := composed.NewStatusPatcherComposed(state.ObjWithStatus())
	if !sp.IsStale() {
		return nil, ctx
	}

	return sp.
		MutateStatus(func(obj composed.ObjWithStatus) {
			cr := obj.(*cloudresourcesv1beta1.CloudResources)
			cr.Status.State = cloudresourcesv1beta1.StateProcessing
			meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
				Type:               cloudresourcesv1beta1.ConditionTypeReady,
				Status:             metav1.ConditionUnknown,
				ObservedGeneration: obj.GetGeneration(),
				Reason:             cloudresourcesv1beta1.ReasonProcessing,
				Message:            cloudresourcesv1beta1.ReasonProcessing,
			})
		}).
		OnSuccess(composed.Continue).
		Run(ctx, state.Cluster().K8sClient())
}
