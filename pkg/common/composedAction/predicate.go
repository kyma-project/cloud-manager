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

type Case interface {
	Predicate(ctx context.Context, state LoggableState) bool
	Action(ctx context.Context, state LoggableState) (*ctrl.Result, error)
}

type CaseStruct struct {
	P Predicate
	A Action
}

func (cs *CaseStruct) Predicate(ctx context.Context, state LoggableState) bool {
	return cs.P(ctx, state)
}

func (cs *CaseStruct) Action(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
	return cs.A(ctx, state)
}

func BuildSwitchAction(name string, defaultAction Action, cases ...Case) Action {
	return func(ctx context.Context, state LoggableState) (result *ctrl.Result, err error) {
		logger := state.Logger().With("action", name)
		for i, cs := range cases {
			value := cs.Predicate(ctx, state)
			if value {
				actionName := findActionName(cs.Action)
				logger.
					With("targetAction", actionName).
					With("index", i).
					Info("Running action")
				return cs.Action(ctx, state)
			}
		}

		if defaultAction != nil {
			actionName := findActionName(defaultAction)
			logger.
				With("targetAction", actionName).
				With("index", "default").
				Info("Running action")
			return defaultAction(ctx, state)
		} else {
			logger.Info("None of case predicates evaluated true, and default action is not provided")
		}

		return nil, nil
	}
}
