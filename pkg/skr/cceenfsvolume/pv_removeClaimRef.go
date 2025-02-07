package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func pvRemoveClaimRef(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}
	if state.PV == nil {
		return nil, ctx
	}
	if state.PV.Status.Phase != corev1.VolumeReleased {
		return nil, nil
	}
	if state.PV.Spec.ClaimRef == nil {
		return nil, nil
	}

	state.PV.Spec.ClaimRef = nil
	err := state.Cluster().K8sClient().Update(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating PV to remove ClaimRef", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx
}
