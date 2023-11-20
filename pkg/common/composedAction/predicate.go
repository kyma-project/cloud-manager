package composed

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Predicate = func(ctx context.Context, state LoggableState) bool

func BuildBranchingAction(name string, predicate Predicate, trueAction Action, falseAction Action) Action {
	return func(ctx context.Context, state LoggableState) (result *ctrl.Result, err error) {
		value := predicate(ctx, state)
		logger := state.Logger().
			With("action", name).
			With("predicate", value)
		if value && trueAction != nil {
			actionName := findActionName(trueAction)
			logger.
				With("targetAction", actionName).
				Info("Running action")
			return trueAction(ctx, state)
		} else if falseAction != nil {
			actionName := findActionName(falseAction)
			logger.
				With("targetAction", actionName).
				Info("Running action")
			return falseAction(ctx, state)
		}

		logger.
			With("targetAction", nil).
			Info("No action called since not supplied")

		return nil, nil
	}
}
