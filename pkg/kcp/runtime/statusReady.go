package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	if !awsProviderPredicate(ctx, st) &&
		!azureProviderPredicate(ctx, st) &&
		!gcpProviderPredicate(ctx, st) {
		// security feature is disabled
		return nil, ctx
	}
	state := st.(*State)

	ds := state.SecurityDesiredState()
	if ds != nil {
		defaultSecurityGate.markSuccess(ds)
	}

	composed.LoggerFromCtx(ctx).Info("Runtime provider success - Ready")

	return state.PatchStatusAnnotations(ctx, "Ready", "Ready", state.ObjAsRuntime().Generation)
}
