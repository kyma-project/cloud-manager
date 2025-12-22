package subscription

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func handleNonGardenerSubscriptionType(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsSubscription().Spec.Details.Garden != nil {
		return nil, ctx
	}

	sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())

	state.ObjAsSubscription().SetStatusReady()

	state.ObjAsSubscription().Status.SubscriptionInfo.Aws = state.ObjAsSubscription().Spec.Details.Aws
	state.ObjAsSubscription().Status.SubscriptionInfo.Azure = state.ObjAsSubscription().Spec.Details.Azure
	state.ObjAsSubscription().Status.SubscriptionInfo.Gcp = state.ObjAsSubscription().Spec.Details.Gcp
	state.ObjAsSubscription().Status.SubscriptionInfo.OpenStack = state.ObjAsSubscription().Spec.Details.Openstack

	return sp.
		OnSuccess(composed.Forget).
		OnFailure(composed.Log("Error patching Subscription status with ready state for non-gardener type")).
		Run(ctx, state.Cluster().K8sClient())
}
