package awsnfsvolume

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

	if state.Volume == nil {
		logger.Info("PersistentVolume for AwsNfsVolume not present.")
		return nil, nil
	}

	if state.Volume.Status.Phase != corev1.VolumeReleased {
		return nil, nil
	}

	if state.Volume.Spec.ClaimRef == nil {
		// waiting for PV controller to update status.Phase to AVAILABLE
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	volumeCopy := state.Volume.DeepCopy()
	volumeCopy.Spec.ClaimRef = nil
	err := state.Cluster().K8sClient().Update(ctx, volumeCopy)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error removing claimRef from PV", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
