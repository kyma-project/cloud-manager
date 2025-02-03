package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func scopeDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	scope := &cloudcontrolv1beta1.Scope{}

	// Load Scope

	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: state.ObjAsNuke().Namespace,
		Name:      state.ObjAsNuke().Spec.Scope.Name,
	}, scope)

	if apierrors.IsNotFound(err) {
		logger.Info("Scope does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, ctx)
	}

	// Delete Scope

	err = state.Cluster().K8sClient().Delete(ctx, scope)
	if apierrors.IsNotFound(err) {
		logger.Error(err, "Error deleting loaded Scope since it does not exist anymore")
	}
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Scope", composed.StopWithRequeue, ctx)
	}

	// Remove finalizer from Scope

	_, err = composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, scope, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patch removing scope finalizer when Nuke has deleted all orphan resources", composed.StopWithRequeue, ctx)
	}

	logger.Info("Scope deleted")

	return nil, ctx
}
