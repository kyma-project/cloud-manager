package v2

import (
	"context"
	"errors"
)

type Action func(ctx context.Context) (context.Context, error)

// ComposedAction ===========================

func ComposeActions(actions ...Action) Action {
	return func(ctx context.Context) (context.Context, error) {
		var lastError error
		currentCtx := ctx
	loop:
		for _, act := range actions {
			select {
			case <-currentCtx.Done():
				lastError = currentCtx.Err()
				break loop
			default:
				nextCtx, err := act(currentCtx)
				lastError = err
				if nextCtx != nil {
					currentCtx = nextCtx
				}
				if err != nil {
					break loop
				}
			}
		}

		var fce FlowControlError
		if lastError == nil {
			return currentCtx, nil
		} else if errors.As(lastError, &fce) {
			if !fce.ShouldReturnError() {
				lastError = nil
			}
			return currentCtx, lastError
		}

		return currentCtx, lastError
	}
}
