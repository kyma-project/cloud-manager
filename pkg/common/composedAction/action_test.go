package composed

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wojas/genericr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
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
	return func(ctx context.Context, state State) error {
		st := state.(*composedActionTestState)
		st.log = append(st.log, logName)
		return err
	}
}

func newComposedActionTestState() *composedActionTestState {
	return &composedActionTestState{
		State: NewState(nil, nil, types.NamespacedName{}, nil),
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

	_ = a(me.ctx, state)
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

	err := a(me.ctx, state)
	assert.Equal(me.T(), []string{"1", "2"}, state.log)
	assert.Equal(me.T(), e1, err)
}

func buildDelayedTestAction(logName string, delay time.Duration, err error) Action {
	return func(ctx context.Context, state State) error {
		st := state.(*composedActionTestState)
		st.log = append(st.log, logName+".start")
		select {
		case <-ctx.Done():
			st.log = append(st.log, logName+".canceled")
		case <-time.After(delay):
			st.log = append(st.log, logName+".end")
		}
		return err
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
		_ = a(ctx, state)
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

func (me *composedActionSuite) TestLogsAllRunActions() {
	myName := "TestLogsAllRunActions"

	var allLogs []genericr.Entry
	ctx := log.IntoContext(me.ctx, logr.New(genericr.New(func(e genericr.Entry) {
		allLogs = append(allLogs, e)
	})))

	state := newComposedActionTestState()

	a := ComposeActions(
		myName,
		buildTestAction("1", nil),
		buildTestAction("2", nil),
		buildTestAction("3", nil),
	)

	_ = a(ctx, state)

	assert.Len(me.T(), allLogs, 4)

	actionName := "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction.(*composedActionSuite).TestLogsAllRunActions.func"

	assert.Equal(me.T(), "Running action", allLogs[0].Message)
	assert.Equal(me.T(), myName, allLogs[0].FieldsMap()["action"])
	assert.Equal(me.T(), actionName+"2", allLogs[0].FieldsMap()["targetAction"])

	assert.Equal(me.T(), "Running action", allLogs[1].Message)
	assert.Equal(me.T(), myName, allLogs[1].FieldsMap()["action"])
	assert.Equal(me.T(), actionName+"3", allLogs[1].FieldsMap()["targetAction"])

	assert.Equal(me.T(), "Running action", allLogs[2].Message)
	assert.Equal(me.T(), myName, allLogs[2].FieldsMap()["action"])
	assert.Equal(me.T(), actionName+"4", allLogs[2].FieldsMap()["targetAction"])

	assert.Equal(me.T(), "Reconciliation finished", allLogs[3].Message)
	assert.Equal(me.T(), myName, allLogs[3].FieldsMap()["action"])
	assert.Equal(me.T(), actionName+"4", allLogs[3].FieldsMap()["lastAction"])
	assert.Equal(me.T(), reconcile.Result{}, allLogs[3].FieldsMap()["result"])
	assert.Equal(me.T(), nil, allLogs[3].FieldsMap()["err"])
}

func TestComposedAction(t *testing.T) {
	suite.Run(t, new(composedActionSuite))
}
