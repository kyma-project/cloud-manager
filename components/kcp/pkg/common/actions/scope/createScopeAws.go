package scope

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func createScopeAws(ctx context.Context, state composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	err := errors.New("aws scope definition not implemented")
	logger.Error(err, "error defining AWS scope")

	return composed.StopAndForget, nil // no requeue
}
