package subscription

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())

	state.ObjAsSubscription().SetStatusDeleting()

	return sp.
		OnSuccess(composed.Continue).
		OnFailure(composed.Log("Failed to patch Subscription status with deleting condition")).
		Run(ctx, state.Cluster().K8sClient())
}
