package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusStaleProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if state.ObjAsObjWithStatus().GetGeneration() != state.ObjAsObjWithStatus().ObservedGeneration() {
		return composed.NewStatusPatcherComposed(state.ObjAsObjWithStatus()).
			MutateStatus(func(o composed.ObjWithStatus) {
				meta.SetStatusCondition(o.Conditions(), metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionUnknown,
					ObservedGeneration: o.GetGeneration(),
					Reason:             cloudcontrolv1beta1.ReasonProcessing,
					Message:            cloudcontrolv1beta1.ReasonProcessing,
				})
			}).
			Run(ctx, state.Cluster().K8sClient())
	}

	return nil, ctx
}
