package composed

import (
	"context"
	"reflect"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func findActionName(a Action) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	return fullName
}

type Action func(ctx context.Context, state State) (error, context.Context)

// ===========================

func ComposeActions(name string, actions ...Action) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		logger := log.FromContext(ctx).WithValues("action", name)
		var actionName string
		var lastError error
		currentCtx := ctx
	loop:
		for _, a := range actions {
			actionName = findActionName(a)
			select {
			case <-currentCtx.Done():
				lastError = currentCtx.Err()
				break loop
			default:
				logger.
					WithValues("targetAction", actionName).
					Info("Running action")
				err, nextCtx := a(currentCtx, state)
				lastError = err
				if err != nil {
					break loop
				}
				if nextCtx != nil {
					currentCtx = nextCtx
				}
			}
		}

		l := logger.
			WithValues(
				"lastAction", actionName,
			)
		if lastError == nil {
			l.Info("Reconciliation finished")
		} else {
			l.Error(lastError, "reconciliation finished")
		}

		return lastError, currentCtx
	}
}
