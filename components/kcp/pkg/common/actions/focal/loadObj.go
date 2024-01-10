package focal

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func loadObj(ctx context.Context, state composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	err := state.LoadObj(ctx)
	if err != nil {
		err = fmt.Errorf("error getting object: %w", err)
		logger.Error(err, "error")
		return composed.StopWithRequeue, nil
	}

	logger.Info("Object loaded")

	return nil, nil
}
