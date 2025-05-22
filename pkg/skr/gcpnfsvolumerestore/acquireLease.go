package gcpnfsvolumerestore

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
)

func acquireLease(ctx context.Context, st composed.State) (error, context.Context) {
	//If deleting, continue with next steps.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsGcpNfsVolumeRestore()
	volumeNamespacedName := restore.Spec.Destination.Volume.ToNamespacedName(restore.Namespace)

	leaseName := getLeaseName(volumeNamespacedName.Name, "restore")
	leaseNamespace := volumeNamespacedName.Namespace
	holderName := getHolderName(types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace})
	leaseDuration := int32(config.SkrRuntimeConfig.SkrLockingLeaseDuration.Seconds())

	res, err := leases.Acquire(
		ctx,
		state.SkrCluster,
		leaseName,
		leaseNamespace,
		holderName,
		leaseDuration,
	)

	switch res {
	case leases.AcquiredLease, leases.RenewedLease:
		return nil, nil
	case leases.LeasingFailed:
		return composed.LogErrorAndReturn(err, "Error acquiring lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	case leases.OtherLeased:
		logger.Info("Another restore leased the filestore. Waiting for it to release the lease.")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	default:
		return composed.LogErrorAndReturn(err, "Unknown lease result", composed.StopAndForget, ctx)
	}
}
