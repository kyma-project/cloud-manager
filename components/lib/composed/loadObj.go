package composed

import (
	"context"
	"fmt"
)

func LoadObj(ctx context.Context, state State) (error, context.Context) {
	logger := LoggerFromCtx(ctx)
	err := state.LoadObj(ctx)
	if err != nil {
		err = fmt.Errorf("error getting object: %w", err)
		logger.Error(err, "error")
		return StopWithRequeue, nil
	}

	logger.Info("Object loaded")

	return nil, nil
}
