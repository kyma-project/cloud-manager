package kyma

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// kymaRemoveStaleCloudManagerFinalizer removes the cloud-manager finalizer that was
// previously added to Kyma CRs but is no longer managed by cloud-manager. The Kyma
// reconciler (KLM) now only removes finalizers it owns, so stale cloud-manager
// finalizers would block Kyma deletion indefinitely.
func kymaRemoveStaleCloudManagerFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !controllerutil.ContainsFinalizer(state.ObjAsKyma(), api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	_, err := state.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error removing stale cloud-manager finalizer from Kyma", composed.StopWithRequeue, ctx)
	}

	composed.LoggerFromCtx(ctx).Info("Removed stale cloud-manager finalizer from Kyma")

	return composed.StopWithRequeue, ctx
}
