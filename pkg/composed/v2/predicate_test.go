package v2

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
		Predicate(func(ctx context.Context) bool {
			return true
		}),
		func(ctx context.Context) (context.Context, error) {
			isTrueActionCalled = true
			return ctx, nil
		},
		func(ctx context.Context) (context.Context, error) {
			assert.Fail(me.T(), "falseAction should not be called")
			return ctx, nil
		},
	)

	_, _ = a(context.Background())

	assert.True(me.T(), isTrueActionCalled)
}

func (me *PredicateSuite) TestRunsFalseAction() {
	isFalseActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		Predicate(func(ctx context.Context) bool {
			return false
		}),
		func(ctx context.Context) (context.Context, error) {
			assert.Fail(me.T(), "trueAction should not be called")
			return ctx, nil
		},
		func(ctx context.Context) (context.Context, error) {
			isFalseActionCalled = true
			return ctx, nil
		},
	)

	_, _ = a(context.Background())

	assert.True(me.T(), isFalseActionCalled)
}

func TestPredicate(t *testing.T) {
	suite.Run(t, new(PredicateSuite))
}
