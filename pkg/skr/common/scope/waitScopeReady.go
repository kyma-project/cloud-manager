package scope

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitScopeReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if state.ObjWithStatus() == nil {
		// unsupported type, likely older API style with object
		// implementing composed.ObjWithConditionsAndState
		return nil, ctx
	}

	readyCond := meta.FindStatusCondition(state.Scope().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if readyCond != nil && readyCond.Status == metav1.ConditionTrue {
		// scope is ready
		return nil, ctx
	}

	return composed.NewStatusPatcherComposed(state.ObjWithStatus()).
		MutateStatus(func(obj composed.ObjWithStatus) {
			meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
				Type:               cloudresourcesv1beta1.ConditionTypeReady,
				Status:             metav1.ConditionUnknown,
				ObservedGeneration: obj.GetGeneration(),
				Reason:             cloudresourcesv1beta1.ReasonWaitingForDependency,
				Message:            "Waiting Scope to become ready",
			})
		}).
		OnStatusChanged(composed.Log("Waiting Scope to become ready")).
		OnSuccess(composed.RequeueAfter(rate.DefaultRegistry.Quick.When(state.Obj()))).
		OnFailure(
			composed.Log(fmt.Sprintf("Error patching status of %T to set condition WaitingForDependency for Scope", state.Obj())),
			composed.RequeueAfter(rate.DefaultRegistry.Quick.When(state.Obj())),
		).
		Run(ctx, state.Cluster().K8sClient())
}
