package composed

import (
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type FlowControlError interface {
	error
	ShouldReturnError() bool
}

type flowControlError struct {
	msg               string
	shouldReturnError bool
}

func (e *flowControlError) Error() string {
	return e.msg
}

func (e *flowControlError) ShouldReturnError() bool {
	return e.shouldReturnError
}

var (
	StopAndForget   error
	StopWithRequeue error
	Break           error
)

func IsFlowControl(err error) bool {
	var eee FlowControlError
	ok := errors.As(err, &eee)
	return ok
}

func IsStopAndForget(err error) bool {
	return errors.Is(err, StopAndForget)
}

func IsStopWithRequeue(err error) bool {
	return errors.Is(err, StopWithRequeue)
}

func IsStopWithRequeueDelay(err error) bool {
	var eee *stopWithRequeueDelay
	ok := errors.As(err, &eee)
	return ok
}

func IsBreak(err error) bool {
	return errors.Is(err, Break)
}

func IsTerminal(err error) bool {
	return errors.Is(err, reconcile.TerminalError(nil))
}

func init() {
	StopAndForget = &flowControlError{
		msg:               "stop and forget",
		shouldReturnError: true,
	}
	StopWithRequeue = &flowControlError{
		msg:               "stop with requeue",
		shouldReturnError: true,
	}
	Break = &flowControlError{
		msg:               "break",
		shouldReturnError: false,
	}
}

type stopWithRequeueDelay struct {
	delay time.Duration
}

func (e *stopWithRequeueDelay) ShouldReturnError() bool {
	return true
}

func (e *stopWithRequeueDelay) Error() string {
	return fmt.Sprintf("stop with requeue delay: %s", e.delay)
}

func (e *stopWithRequeueDelay) Delay() time.Duration {
	return e.delay
}

func StopWithRequeueDelay(d time.Duration) error {
	return &stopWithRequeueDelay{delay: d}
}

func AnyConditionChanged(obj ObjWithConditions, conditionsToSet ...metav1.Condition) bool {
	return pie.All(conditionsToSet, func(x metav1.Condition) bool {
		c := meta.FindStatusCondition(*obj.Conditions(), x.Type)
		return c == nil || c.Reason != x.Reason || c.Message != x.Message || c.Status != x.Status
	})
}

func SyncConditions(obj ObjWithConditions, conditionsToSet ...metav1.Condition) bool {
	conditionsToRemove := pie.Filter(*obj.Conditions(), func(x metav1.Condition) bool {
		return meta.FindStatusCondition(conditionsToSet, x.Type) == nil
	})

	changed := false
	for _, condition := range conditionsToRemove {
		if meta.RemoveStatusCondition(obj.Conditions(), condition.Type) {
			changed = true
		}
	}

	for _, condition := range conditionsToSet {
		if meta.SetStatusCondition(obj.Conditions(), condition) {
			changed = true
		}
	}
	return changed
}
