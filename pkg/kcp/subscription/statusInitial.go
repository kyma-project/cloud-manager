package subscription

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitial(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())
	if !sp.IsStale() {
		return nil, ctx
	}

	return sp.
		MutateStatus(func(obj *cloudcontrolv1beta1.Subscription) {
			obj.SetStatusProcessing()
		}).
		OnSuccess(composed.Continue).
		OnFailure(composed.Log("Error setting initial status on Subscription")).
		Run(ctx, state.Cluster().K8sClient())
}
