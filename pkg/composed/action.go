package composed

import (
	"context"
	"errors"
	"reflect"
	"runtime"
)

func findActionName(a Action) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	return fullName
}

type Action func(ctx context.Context, state State) (error, context.Context)

func Noop(ctx context.Context, state State) (error, context.Context) {
	return nil, ctx
}

// ===========================

func ComposeActionsNoName(actions ...Action) Action {
	return ComposeActions("", actions...)
}

func ComposeActions(_ string, actions ...Action) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		var lastError error
		currentCtx := ctx
	loop:
		for _, a := range actions {
			select {
			case <-currentCtx.Done():
				lastError = currentCtx.Err()
				break loop
			default:
				err, nextCtx := a(currentCtx, state)
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
			return nil, currentCtx
		} else if errors.As(lastError, &fce) {
			if !fce.ShouldReturnError() {
				lastError = nil
			}
			return lastError, currentCtx
		}

		return lastError, currentCtx
	}
}
