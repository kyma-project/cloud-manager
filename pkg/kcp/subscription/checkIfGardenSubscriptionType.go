package subscription

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkIfGardenSubscriptionType(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsSubscription().Spec.Details.Garden == nil {
		return composed.NewStatusPatcherComposed(state.ObjAsSubscription()).
			MutateStatus(func(obj *cloudcontrolv1beta1.Subscription) {
				obj.SetStatusInvalidSpec("Unknown subscription type")
			}).
			OnStatusChanged(
				composed.Log("Unable to handle Subscription type"),
			).
			OnSuccess(composed.Forget).
			OnFailure(composed.Log("Error patching Subscription status with invalid type")).
			Run(ctx, state.Cluster().K8sClient())
	}

	return nil, ctx
}
