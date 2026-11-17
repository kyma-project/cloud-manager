package scope

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if state.Scope() != nil {
		return nil, ctx
	}

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.KcpCluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef().Namespace,
		Name:      state.KymaRef().Name,
	}, scope)
	if apierrors.IsNotFound(err) {
		// new, preferred, style
		if state.ObjWithStatus() != nil {
			return composed.NewStatusPatcherComposed(state.ObjWithStatus()).
				MutateStatus(func(obj composed.ObjWithStatus) {
					meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
						Type:               cloudcontrolv1beta1.ConditionTypeReady,
						Status:             metav1.ConditionFalse,
						ObservedGeneration: obj.GetGeneration(),
						Reason:             cloudresourcesv1beta1.ConditionReasonMissingScope,
						Message:            "Scope for SKR not found",
					})
				}).
				OnStatusChanged(composed.Log("Scope for SKR not found")).
				OnSuccess(composed.RequeueAfter(util.Timing.T10000ms())).
				Run(ctx, state.Cluster().K8sClient())
		} else if state.ObjWithConditionsAndState() != nil {
			// legacy style
			state.ObjWithConditionsAndState().SetState(cloudresourcesv1beta1.StateError)
			return composed.UpdateStatus(state.ObjWithConditionsAndState()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonMissingScope,
					Message: "Scope for SKR not found",
				}).
				ErrorLogMessage("Error updating object status after setting missing Scope status").
				SuccessLogMsg("Forgetting object due to missing Scope").
				Run(ctx, state)
		}

	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP Scope", composed.StopWithRequeue, ctx)
	}

	state.SetScope(scope)
	return nil, ctx
}
