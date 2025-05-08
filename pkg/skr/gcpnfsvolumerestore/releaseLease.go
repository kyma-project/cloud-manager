package gcpnfsvolumerestore

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func releaseLease(ctx context.Context, st composed.State) (error, context.Context) {
	//If it reaches here, the operation is in final status
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolumeRestore()
	volumeNamespacedName := restore.Spec.Destination.Volume.ToNamespacedName(restore.Namespace)

	leaseName := getLeaseName(volumeNamespacedName.Name, "restore")
	leaseNamespace := volumeNamespacedName.Namespace
	holderName := getHolderName(types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace})

	err := leases.Release(
		ctx,
		state.SkrCluster,
		leaseName,
		leaseNamespace,
		holderName,
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error releasing lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	}
	return nil, nil
}
