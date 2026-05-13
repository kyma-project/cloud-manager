package sapnfsvolumesnapshotrestore

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func releaseLease(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	destRef := restore.Spec.Destination.ExistingVolume
	if destRef == nil {
		// New-volume restore, no lease to release
		return nil, ctx
	}

	ns := destRef.Namespace
	if ns == "" {
		ns = restore.Namespace
	}

	leaseName := getLeaseName(destRef.Name)
	leaseNamespace := ns
	holderName := getHolderName(types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace})

	err := leases.Release(
		ctx,
		state.SkrCluster,
		leaseName,
		leaseNamespace,
		holderName,
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ctx
		}
		if strings.Contains(err.Error(), "belongs to another owner") {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error releasing lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	}
	return nil, ctx
}
