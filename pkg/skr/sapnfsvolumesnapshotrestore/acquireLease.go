package sapnfsvolumesnapshotrestore

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntimeconfig "github.com/kyma-project/cloud-manager/pkg/skr/runtime/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
)

func acquireLease(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	destRef := restore.Spec.Destination.ExistingVolume
	ns := destRef.Namespace
	if ns == "" {
		ns = restore.Namespace
	}

	leaseName := getLeaseName(destRef.Name)
	leaseNamespace := ns
	holderName := getHolderName(types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace})
	leaseDuration := int32(skrruntimeconfig.SkrRuntimeConfig.SkrLockingLeaseDuration.Seconds())

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
		return nil, ctx
	case leases.LeasingFailed:
		return composed.LogErrorAndReturn(err, "Error acquiring lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	case leases.OtherLeased:
		logger.Info("Another restore holds the lease on the volume. Waiting for it to release.")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	default:
		return composed.LogErrorAndReturn(err, "Unknown lease result", composed.StopAndForget, ctx)
	}
}
