package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// If PV is in status.Phase RELEASED, spec.claimRef has to be removed
// so the PV can transfer to status.Phase AVAILABLE
// and become ready to be attached to PVC
func sanitizeReleasedVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.PV == nil {
		logger.Info("PersistentVolume for GcpNfsVolume not present.")
		return nil, nil
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
		return composed.LogErrorAndReturn(err, "Error removing claimRef from PV", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
