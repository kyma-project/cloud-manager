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
		func(ctx context.Context, state State) bool {
			return true
		},
		func(ctx context.Context, state State) error {
			isTrueActionCalled = true
			return nil
		},
		func(ctx context.Context, state State) error {
			assert.Fail(me.T(), "falseAction should not be called")
			return nil
		},
	)
	state := newComposedActionTestState()

	_ = a(context.Background(), state)

	assert.True(me.T(), isTrueActionCalled)
}

func (me *PredicateSuite) TestRunsFalseAction() {
	isFalseActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		func(ctx context.Context, state State) bool {
			return false
		},
		func(ctx context.Context, state State) error {
			assert.Fail(me.T(), "trueAction should not be called")
			return nil
		},
		func(ctx context.Context, state State) error {
			isFalseActionCalled = true
			return nil
		},
	)
	state := newComposedActionTestState()

	_ = a(context.Background(), state)

	assert.True(me.T(), isFalseActionCalled)
}

func TestPredicate(t *testing.T) {
	suite.Run(t, new(PredicateSuite))
}
