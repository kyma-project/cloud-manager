package composed

import (
	"context"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/record"
	"reflect"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoggableState interface {
	Logger() *zap.SugaredLogger
}

type K8sState interface {
	LoggableState
	client.Client
	record.EventRecorder
}

func findActionName(a Action) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	return fullName
}

type Action = func(ctx context.Context, state LoggableState) (*ctrl.Result, error)

// ===========================

func ComposeActions(name string, actions ...Action) Action {
	return func(ctx context.Context, state LoggableState) (result *ctrl.Result, err error) {
		logger := state.Logger().With("action", name)
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
					With("targetAction", actionName).
					Info("Running action")
				result, err = a(ctx, state)
				if err != nil {
					break loop
				}
			}
		}

		l := logger.
			With("lastAction", actionName).
			With("result", result).
			With("err", err)
		if err == nil {
			l.Info("Reconciliation finished")
		} else {
			l.Error("reconciliation finished")
		}

		return result, err
	}
}
