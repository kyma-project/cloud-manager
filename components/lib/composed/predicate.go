package composed

import (
	"context"
)

type Predicate func(ctx context.Context, state State) bool

func Not(p Predicate) Predicate {
	return func(ctx context.Context, state State) bool {
		return !p(ctx, state)
	}
}

func All(predicates ...Predicate) Predicate {
	return func(ctx context.Context, state State) bool {
		for _, p := range predicates {
			if !p(ctx, state) {
				return false
			}
		}
		return true
	}
}

func Any(predicates ...Predicate) Predicate {
	return func(ctx context.Context, state State) bool {
		for _, p := range predicates {
			if p(ctx, state) {
				return true
			}
		}
		return false
	}
}

func BuildBranchingAction(name string, predicate Predicate, trueAction Action, falseAction Action) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		value := predicate(ctx, state)
		logger := LoggerFromCtx(ctx).
			WithValues(
				"action", name,
				"predicate", value,
			)
		if value && trueAction != nil {
			actionName := findActionName(trueAction)
			logger.
				WithValues("targetAction", actionName).
				Info("Running action")
			return trueAction(ctx, state)
		} else if falseAction != nil {
			actionName := findActionName(falseAction)
			logger.
				WithValues("targetAction", actionName).
				Info("Running action")
			return falseAction(ctx, state)
		}

		logger.
			WithValues("targetAction", nil).
			Info("No action called since not supplied")

		return nil, nil
	}
}

type Case interface {
	Predicate(ctx context.Context, state State) bool
	Action(ctx context.Context, state State) (error, context.Context)
}

func NewCase(p Predicate, a Action) Case {
	return &CaseStruct{
		P: p,
		A: a,
	}
}

type CaseStruct struct {
	P Predicate
	A Action
}

func (cs *CaseStruct) Predicate(ctx context.Context, state State) bool {
	return cs.P(ctx, state)
}

func (cs *CaseStruct) Action(ctx context.Context, state State) (error, context.Context) {
	return cs.A(ctx, state)
}

func BuildSwitchAction(name string, defaultAction Action, cases ...Case) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		logger := LoggerFromCtx(ctx).WithValues("action", name)
		for i, cs := range cases {
			value := cs.Predicate(ctx, state)
			if value {
				actionName := findActionName(cs.Action)
				logger.
					WithValues(
						"targetAction", actionName,
						"index", i,
					).
					Info("Running action")
				return cs.Action(ctx, state)
			}
		}

		if defaultAction != nil {
			actionName := findActionName(defaultAction)
			logger.
				WithValues(
					"targetAction", actionName,
					"index", "default",
				).
				Info("Running action")
			return defaultAction(ctx, state)
		} else {
			logger.Info("None of case predicates evaluated true, and default action is not provided")
		}

		return nil, nil
	}
}
