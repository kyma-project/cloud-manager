package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func pvLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	vol := &corev1.PersistentVolume{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Name: state.ObjAsCceeNfsVolume().GetPVName(),
	}, vol)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading PV for CceeNfsVolume", composed.StopWithRequeue, ctx)
	}

	state.PV = vol

	return nil, ctx
}
