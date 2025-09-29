package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		state.ObjWithConditionsAndState().SetState(cloudresourcesv1beta1.StateError)
		return composed.UpdateStatus(state.ObjWithConditionsAndState()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingScope,
				Message: "Scope for SKR not found",
			}).
			ErrorLogMessage("Error updating object status after setting missing Scope status").
			SuccessLogMsg("Forgeting object due to missing Scope").
			Run(ctx, state)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP Scope", composed.StopWithRequeue, ctx)
	}

	state.SetScope(scope)
	return nil, ctx
}
