package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	gcpNfsVolume := state.ObjAsGcpNfsVolume()

	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      getVolumeClaimName(gcpNfsVolume),
	}, pvc)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolumeClaim by getVolumeName()", composed.StopWithRequeue, ctx)
	}

	if err != nil { // PVC not-found
		return nil, nil
	}

	state.PVC = pvc

	return nil, nil
}
