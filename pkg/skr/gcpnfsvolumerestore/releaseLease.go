package gcpnfsvolumerestore

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/leases"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func releaseLease(ctx context.Context, st composed.State) (error, context.Context) {
	//If it reaches here, the operation is in final status
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolumeRestore()
	err := leases.Release(ctx, state.SkrCluster,
		types.NamespacedName{Name: state.GcpNfsVolume.Name, Namespace: state.GcpNfsVolume.Namespace},
		types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace},
		"restore")
	if err != nil && !apierrors.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error releasing lease", composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx)
	}
	return nil, nil
}
