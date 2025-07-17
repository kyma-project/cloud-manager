package sapnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func pvcLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: state.Obj().GetNamespace(),
		Name:      state.ObjAsSapNfsVolume().GetPVCName(),
	}, pvc)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading PVC for SapNfsVolume", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.PVC = pvc
	}

	return nil, ctx
}
