package kyma

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func scopeLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKeyFromObject(state.Obj()), scope)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading scope for kyma", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.scope = scope
		logger := composed.LoggerFromCtx(ctx)
		logger = logger.WithValues("scope-loaded", "true")
		ctx = composed.LoggerIntoCtx(ctx, logger)
	}

	return nil, ctx
}
