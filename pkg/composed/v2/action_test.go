package v2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type composedActionSuite struct {
	suite.Suite
	ctx context.Context
}

type composedActionTestState struct {
	State
	log []string
}

func buildTestAction(logName string, err error) Action {
	return func(ctx context.Context) (context.Context, error) {
		state := StateFromCtx[*composedActionTestState](ctx)
		state.log = append(state.log, logName)
		return ctx, err
	}
}

func newComposedActionTestState() *composedActionTestState {
	return &composedActionTestState{}
}

func (me *composedActionSuite) SetupTest() {
	me.ctx = log.IntoContext(context.Background(), logr.Discard())
	me.ctx = StateToCtx(me.ctx, newComposedActionTestState())
}

func (me *composedActionSuite) TestAllActionsAreRunInSequence() {

	a := ComposeActions(
		buildTestAction("1", nil),
		buildTestAction("2", nil),
		buildTestAction("3", nil),
	)

	ctx, _ := a(me.ctx)
	state := StateFromCtx[*composedActionTestState](ctx)
	assert.Equal(me.T(), []string{"1", "2", "3"}, state.log)
}

func (me *composedActionSuite) TestBreaksOnFirstError() {
	e1 := errors.New("dummy")
	a := ComposeActions(
		buildTestAction("1", nil),
		buildTestAction("2", e1),
		buildTestAction("3", nil),
	)

	ctx, err := a(me.ctx)
	state := StateFromCtx[*composedActionTestState](ctx)
	assert.Equal(me.T(), []string{"1", "2"}, state.log)
	assert.Equal(me.T(), e1, err)
}

func buildDelayedTestAction(logName string, delay time.Duration, err error) Action {
	return func(ctx context.Context) (context.Context, error) {
		state := StateFromCtx[*composedActionTestState](ctx)
		state.log = append(state.log, logName+".start")
		select {
		case <-ctx.Done():
			state.log = append(state.log, logName+".canceled")
		case <-time.After(delay):
			state.log = append(state.log, logName+".end")
		}
		return ctx, err
	}
}

func (me *composedActionSuite) TestCanBeInterrupted() {
	a := ComposeActions(
		buildDelayedTestAction("1", time.Second, nil),
		buildDelayedTestAction("2", time.Second, nil),
		buildDelayedTestAction("3", time.Second, nil),
	)

	ctx, cancel := context.WithCancel(me.ctx)

	state := StateFromCtx[*composedActionTestState](ctx)

	hasReturned := false
	go func() {
		_, _ = a(ctx)
		hasReturned = true
	}()

	timeout := time.After(3100 * time.Millisecond)
loop:
	for {
		select {
		case <-timeout:
			assert.Fail(me.T(), "timeout")
			cancel()
			return
		default:
			if len(state.log) >= 3 {
				cancel()
				break loop
			}
		}
	}

	time.Sleep(100 * time.Millisecond)

	assert.True(me.T(), hasReturned)
	assert.Len(me.T(), state.log, 4)
	assert.Equal(me.T(), "2.canceled", state.log[3])
}

func TestComposedAction(t *testing.T) {
	suite.Run(t, new(composedActionSuite))
}
