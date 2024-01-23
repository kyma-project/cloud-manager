package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func loadScopeObj(ctx context.Context, state composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	err := state.LoadObj(ctx)
	if apierrors.IsNotFound(err) {
		logger.Info("Scope object does not exist")
		// continue to create one
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope object", composed.StopWithRequeue, nil)
	}

	return composed.StopAndForget, nil
}
