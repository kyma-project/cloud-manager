package composed

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func LoadObj(ctx context.Context, state State) (error, context.Context) {
	logger := LoggerFromCtx(ctx)
	err := state.LoadObj(ctx)
	if apierrors.IsNotFound(err) {
		return StopAndForget, nil
	}
	if err != nil {
		err = fmt.Errorf("error loading object: %w", err)
		logger.Error(err, "error")
		return StopWithRequeue, nil
	}

	return nil, nil
}
