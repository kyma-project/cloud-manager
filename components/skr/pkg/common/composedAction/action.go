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

type Action = func(ctx context.Context, state State) error

// ===========================

func ComposeActions(name string, actions ...Action) Action {
	return func(ctx context.Context, state State) (err error) {
		logger := log.FromContext(ctx).WithValues("action", name)
		var actionName string
	loop:
		for _, a := range actions {
			actionName = findActionName(a)
			select {
			case <-ctx.Done():
				err = ctx.Err()
				break loop
			default:
				logger.
					WithValues("targetAction", actionName).
					Info("Running action")
				err = a(ctx, state)
				if err != nil || state.IsStopped() {
					break loop
				}
			}
		}

		l := logger.
			WithValues(
				"lastAction", actionName,
				"result", state.Result(),
			)
		if err == nil {
			l.Info("Reconciliation finished")
		} else {
			l.Error(err, "reconciliation finished")
		}

		return err
	}
}
