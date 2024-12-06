package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kymaNetworkReferenceDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.kcpNetworkKyma == nil {
		return nil, ctx
	}
	if !state.kcpNetworkKyma.DeletionTimestamp.IsZero() {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Deleting Kyma Network Reference")

	err := state.Cluster().K8sClient().Delete(ctx, state.kcpNetworkKyma)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP kyma Network reference", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, ctx
}
