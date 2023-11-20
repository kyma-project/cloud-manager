package composed

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

type composedActionSuite struct {
	suite.Suite
}

type composedActionTestState struct {
	BaseState
	log []string
}

func buildTestAction(logName string, result *ctrl.Result, err error) Action {
	return func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
		st := state.(*composedActionTestState)
		st.log = append(st.log, logName)
		return result, err
	}
}

func newComposedActionTestState(logger *zap.SugaredLogger) *composedActionTestState {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &composedActionTestState{
		BaseState: BaseState{
			logger: logger,
		},
	}
}

func (me *composedActionSuite) TestAllActionsAreRunInSequence() {
	state := newComposedActionTestState(nil)

	a := ComposeActions(
		"TestAllActionsAreRunInSequence",
		buildTestAction("1", nil, nil),
		buildTestAction("2", nil, nil),
		buildTestAction("3", nil, nil),
	)

	_, _ = a(context.Background(), state)
	assert.Equal(me.T(), []string{"1", "2", "3"}, state.log)
}

func (me *composedActionSuite) TestBreaksOnFirstError() {
	state := newComposedActionTestState(nil)

	e1 := errors.New("dummy")
	a := ComposeActions(
		"TestBreaksOnFirstError",
		buildTestAction("1", nil, nil),
		buildTestAction("2", nil, e1),
		buildTestAction("3", nil, nil),
	)

	_, err := a(context.Background(), state)
	assert.Equal(me.T(), []string{"1", "2"}, state.log)
	assert.Equal(me.T(), e1, err)
}

func buildDelayedTestAction(logName string, delay time.Duration, result *ctrl.Result, err error) Action {
	return func(ctx context.Context, state LoggableState) (*ctrl.Result, error) {
		st := state.(*composedActionTestState)
		st.log = append(st.log, logName+".start")
		select {
		case <-ctx.Done():
			st.log = append(st.log, logName+".canceled")
		case <-time.After(delay):
			st.log = append(st.log, logName+".end")
		}
		return result, err
	}
}

func (me *composedActionSuite) TestCanBeInterrupted() {
	a := ComposeActions(
		"TestCanBeInterrupted",
		buildDelayedTestAction("1", time.Second, nil, nil),
		buildDelayedTestAction("2", time.Second, nil, nil),
		buildDelayedTestAction("3", time.Second, nil, nil),
	)

	state := newComposedActionTestState(nil)
	ctx, cancel := context.WithCancel(context.Background())

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

func (me *composedActionSuite) TestLogsAllRunnedActions() {
	myName := "TestLogsAllRunnedActions"
	observedZapCore, observedLogs := observer.New(zap.InfoLevel)
	observedLogger := zap.New(observedZapCore)

	state := newComposedActionTestState(observedLogger.Sugar())

	a := ComposeActions(
		myName,
		buildTestAction("1", nil, nil),
		buildTestAction("2", nil, nil),
		buildTestAction("3", nil, nil),
	)

	_, _ = a(context.Background(), state)

	allLogs := observedLogs.All()
	assert.Len(me.T(), allLogs, 4)

	actionName := "github.com/kyma-project/redis-manager/pkg/composed.(*composedActionSuite).TestLogsAllRunnedActions.func"

	assert.Equal(me.T(), "Running action", allLogs[0].Message)
	assert.Equal(me.T(), myName, allLogs[0].ContextMap()["action"])
	assert.Equal(me.T(), actionName+"1", allLogs[0].ContextMap()["targetAction"])

	assert.Equal(me.T(), "Running action", allLogs[1].Message)
	assert.Equal(me.T(), myName, allLogs[1].ContextMap()["action"])
	assert.Equal(me.T(), actionName+"2", allLogs[1].ContextMap()["targetAction"])

	assert.Equal(me.T(), "Running action", allLogs[2].Message)
	assert.Equal(me.T(), myName, allLogs[2].ContextMap()["action"])
	assert.Equal(me.T(), actionName+"3", allLogs[2].ContextMap()["targetAction"])

	assert.Equal(me.T(), "Reconciliation finished", allLogs[3].Message)
	assert.Equal(me.T(), myName, allLogs[3].ContextMap()["action"])
	assert.Equal(me.T(), actionName+"3", allLogs[3].ContextMap()["lastAction"])
	assert.Equal(me.T(), (*reconcile.Result)(nil), allLogs[3].ContextMap()["result"])
	assert.Equal(me.T(), nil, allLogs[3].ContextMap()["err"])
}

func TestComposedAction(t *testing.T) {
	suite.Run(t, new(composedActionSuite))
}
