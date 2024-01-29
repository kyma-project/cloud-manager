package composed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
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
	return func(ctx context.Context, st State) (error, context.Context) {
		state := st.(*composedActionTestState)
		state.log = append(state.log, logName)
		return err, nil
	}
}

func newComposedActionTestState() *composedActionTestState {
	return &composedActionTestState{
		State: NewStateFactory(nil).NewState(types.NamespacedName{}, nil),
	}
}

func (me *composedActionSuite) SetupTest() {
	me.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (me *composedActionSuite) TestAllActionsAreRunInSequence() {
	state := newComposedActionTestState()

	a := ComposeActions(
		"TestAllActionsAreRunInSequence",
		buildTestAction("1", nil),
		buildTestAction("2", nil),
		buildTestAction("3", nil),
	)

	_, _ = a(me.ctx, state)
	assert.Equal(me.T(), []string{"1", "2", "3"}, state.log)
}

func (me *composedActionSuite) TestBreaksOnFirstError() {
	state := newComposedActionTestState()

	e1 := errors.New("dummy")
	a := ComposeActions(
		"TestBreaksOnFirstError",
		buildTestAction("1", nil),
		buildTestAction("2", e1),
		buildTestAction("3", nil),
	)

	err, _ := a(me.ctx, state)
	assert.Equal(me.T(), []string{"1", "2"}, state.log)
	assert.Equal(me.T(), e1, err)
}

func buildDelayedTestAction(logName string, delay time.Duration, err error) Action {
	return func(ctx context.Context, st State) (error, context.Context) {
		state := st.(*composedActionTestState)
		state.log = append(state.log, logName+".start")
		select {
		case <-ctx.Done():
			state.log = append(state.log, logName+".canceled")
		case <-time.After(delay):
			state.log = append(state.log, logName+".end")
		}
		return err, nil
	}
}

func (me *composedActionSuite) TestCanBeInterrupted() {
	a := ComposeActions(
		"TestCanBeInterrupted",
		buildDelayedTestAction("1", time.Second, nil),
		buildDelayedTestAction("2", time.Second, nil),
		buildDelayedTestAction("3", time.Second, nil),
	)

	state := newComposedActionTestState()
	ctx, cancel := context.WithCancel(me.ctx)

	hasReturned := false
	go func() {
		_, _ = a(ctx, state)
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
