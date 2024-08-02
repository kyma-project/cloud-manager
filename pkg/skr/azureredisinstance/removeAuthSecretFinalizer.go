package azureredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeAuthSecretFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.AuthSecret == nil {
		return nil, nil
	}

	if !controllerutil.ContainsFinalizer(state.AuthSecret, v1beta1.Finalizer) {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.AuthSecret, v1beta1.Finalizer)
	err := state.Cluster().K8sClient().Update(ctx, state.AuthSecret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR Secret after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
