package composed

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PredicateSuite struct {
	suite.Suite
}

func (me *PredicateSuite) TestRunsTrueAction() {
	isTrueActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		Predicate(func(ctx context.Context, state State) bool {
			return true
		}),
		func(ctx context.Context, state State) (error, context.Context) {
			isTrueActionCalled = true
			return nil, nil
		},
		func(ctx context.Context, state State) (error, context.Context) {
			assert.Fail(me.T(), "falseAction should not be called")
			return nil, nil
		},
	)
	state := newComposedActionTestState()

	_, _ = a(context.Background(), state)

	assert.True(me.T(), isTrueActionCalled)
}

func (me *PredicateSuite) TestRunsFalseAction() {
	isFalseActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		Predicate(func(ctx context.Context, state State) bool {
			return false
		}),
		func(ctx context.Context, state State) (error, context.Context) {
			assert.Fail(me.T(), "trueAction should not be called")
			return nil, nil
		},
		func(ctx context.Context, state State) (error, context.Context) {
			isFalseActionCalled = true
			return nil, nil
		},
	)
	state := newComposedActionTestState()

	_, _ = a(context.Background(), state)

	assert.True(me.T(), isFalseActionCalled)
}

func TestPredicate(t *testing.T) {
	suite.Run(t, new(PredicateSuite))
}
