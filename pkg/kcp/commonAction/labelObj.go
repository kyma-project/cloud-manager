package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func labelObj(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if state.subscription == nil {
		return nil, ctx
	}

	added, err := composed.PatchObjMergeLabels(ctx, cloudcontrolv1beta1.SubscriptionLabel, state.subscription.Name, state.Obj(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error setting subscription label", composed.StopWithRequeue, ctx)
	}

	if added {
		return composed.StopWithRequeue, nil
	}

	return nil, ctx
}

