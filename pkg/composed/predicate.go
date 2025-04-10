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

// All returns a Predicate composed of many given predicates that
// returns true only if all given predicates return true
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

// Any returns a Predicate composed of many given predicates
// that returns true if any of the given predicates returns true
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

func BreakIf(predicate Predicate) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		val := predicate(ctx, state)
		if val {
			return Break, nil
		}
		return nil, nil
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

		return nil, nil
	}
}

type Case interface {
	Predicate(ctx context.Context, state State) bool
	Action(ctx context.Context, state State) (error, context.Context)
}

func NewCase(p Predicate, actions ...Action) Case {
	var a Action
	if len(actions) == 1 {
		a = actions[0]
	} else {
		a = ComposeActions("", actions...)
	}
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

func If(condition Predicate, actions ...Action) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		if condition(ctx, state) {
			return ComposeActions("if", actions...)(ctx, state)
		}
		return nil, nil
	}
}

func IfElse(condition Predicate, trueAction Action, falseAction Action) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		if condition(ctx, state) {
			if trueAction != nil {
				return trueAction(ctx, state)
			}
		} else {
			if falseAction != nil {
				return falseAction(ctx, state)
			}
		}
		return nil, nil
	}
}

func Switch(defaultAction Action, cases ...Case) Action {
	return func(ctx context.Context, state State) (error, context.Context) {
		for _, cs := range cases {
			value := cs.Predicate(ctx, state)
			if value {
				return cs.Action(ctx, state)
			}
		}

		if defaultAction != nil {
			return defaultAction(ctx, state)
		}

		return nil, ctx
	}
}

func BuildSwitchAction(_ string, defaultAction Action, cases ...Case) Action {
	return Switch(defaultAction, cases...)
}
