package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNfsInstanceDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.ObjAsCceeNfsVolume()) {
		return nil, ctx
	}

	if state.KcpNfsInstance == nil {
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(state.KcpNfsInstance) {
		return nil, ctx
	}

	logger.Info("Deleting KCP NfsInstance for CceeNfsVolume")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpNfsInstance)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP NfsInstance for CceeNfsVolume", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
