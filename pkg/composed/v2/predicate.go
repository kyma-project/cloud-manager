package v2

import (
	"context"
)

type Predicate func(ctx context.Context) bool

func Not(p Predicate) Predicate {
	return func(ctx context.Context) bool {
		return !p(ctx)
	}
}

func IsLoaded(ctx context.Context) bool {
	state := StateFromCtx[State](ctx)
	return state.Obj() != nil && state.Obj().GetName() != "" && state.Obj().GetGeneration() > 0
}

// All returns a Predicate composed of many given predicates that
// returns true only if all given predicates return true
func All(predicates ...Predicate) Predicate {
	return func(ctx context.Context) bool {
		for _, p := range predicates {
			if !p(ctx) {
				return false
			}
		}
		return true
	}
}

// Any returns a Predicate composed of many given predicates
// that returns true if any of the given predicates returns true
func Any(predicates ...Predicate) Predicate {
	return func(ctx context.Context) bool {
		for _, p := range predicates {
			if p(ctx) {
				return true
			}
		}
		return false
	}
}

func BreakIf(predicate Predicate) Action {
	return func(ctx context.Context) (context.Context, error) {
		val := predicate(ctx)
		if val {
			return ctx, Break
		}
		return ctx, nil
	}
}

func BuildBranchingAction(name string, predicate Predicate, trueAction Action, falseAction Action) Action {
	return func(ctx context.Context) (context.Context, error) {
		value := predicate(ctx)
		if value && trueAction != nil {
			return trueAction(ctx)
		} else if falseAction != nil {
			return falseAction(ctx)
		}

		return ctx, nil
	}
}

type Case interface {
	Predicate(ctx context.Context) bool
	Action(ctx context.Context) (context.Context, error)
}

func NewCase(p Predicate, actions ...Action) Case {
	var a Action
	if len(actions) == 1 {
		a = actions[0]
	} else {
		a = ComposeActions(actions...)
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

func (cs *CaseStruct) Predicate(ctx context.Context) bool {
	return cs.P(ctx)
}

func (cs *CaseStruct) Action(ctx context.Context) (context.Context, error) {
	return cs.A(ctx)
}

func If(condition Predicate, actions ...Action) Action {
	return func(ctx context.Context) (context.Context, error) {
		if condition(ctx) {
			return ComposeActions(actions...)(ctx)
		}
		return ctx, nil
	}
}

func IfElse(condition Predicate, trueAction Action, falseAction Action) Action {
	return func(ctx context.Context) (context.Context, error) {
		if condition(ctx) {
			if trueAction != nil {
				return trueAction(ctx)
			}
		} else {
			if falseAction != nil {
				return falseAction(ctx)
			}
		}
		return ctx, nil
	}
}

func BuildSwitchAction(name string, defaultAction Action, cases ...Case) Action {
	return func(ctx context.Context) (context.Context, error) {
		for _, cs := range cases {
			value := cs.Predicate(ctx)
			if value {
				return cs.Action(ctx)
			}
		}

		if defaultAction != nil {
			return defaultAction(ctx)
		}

		return ctx, nil
	}
}
