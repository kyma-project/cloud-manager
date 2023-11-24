package composed

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

type PredicateSuite struct {
	suite.Suite
}

func (me *PredicateSuite) TestRunsTrueAction() {
	isTrueActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		func(ctx context.Context, state LoggableState) bool {
			return true
		},
		func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
			isTrueActionCalled = true
			return nil, nil
		},
		func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
			assert.Fail(me.T(), "falseAction should not be called")
			return nil, nil
		},
	)
	state := newComposedActionTestState(nil)

	_, _ = a(context.Background(), state)

	assert.True(me.T(), isTrueActionCalled)
}

func (me *PredicateSuite) TestRunsFalseAction() {
	isFalseActionCalled := false
	a := BuildBranchingAction(
		"runsTrue",
		func(ctx context.Context, state LoggableState) bool {
			return false
		},
		func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
			assert.Fail(me.T(), "trueAction should not be called")
			return nil, nil
		},
		func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
			isFalseActionCalled = true
			return nil, nil
		},
	)
	state := newComposedActionTestState(nil)

	_, _ = a(context.Background(), state)

	assert.True(me.T(), isFalseActionCalled)
}

func TestPredicate(t *testing.T) {
	suite.Run(t, new(PredicateSuite))
}
