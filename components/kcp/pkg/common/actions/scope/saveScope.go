package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func saveScope(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	err := state.Client().Create(ctx, state.Scope())
	if err != nil {
		// it's possible that some other loop concurrently running already created this
		// scope in the meanwhile since we checked if it exists in this loop
		// requeue so it tries to find it again now that it exists
		// or, just a glitch, also retry
		err = fmt.Errorf("error creating scope: %w", err)
		logger.Error(err, "error saving scope")
		return composed.StopWithRequeue, nil // will requeue
	}

	return nil, nil
}
