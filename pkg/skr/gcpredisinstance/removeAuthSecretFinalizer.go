package gcpredisinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeAuthSecretFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.AuthSecret == nil {
		return nil, nil
	}

	if !controllerutil.ContainsFinalizer(state.AuthSecret, api.CommonFinalizerDeletionHook) {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.AuthSecret, api.CommonFinalizerDeletionHook)
	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR Secret after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
