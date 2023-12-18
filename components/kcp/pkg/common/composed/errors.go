package composed

import (
	"errors"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	StopAndForget   error
	StopWithRequeue error
)

func IsStopAndForget(err error) bool {
	return err == StopAndForget
}

func IsStopWithRequeue(err error) bool {
	return err == StopWithRequeue
}

func IsStopWithRequeueDelay(err error) bool {
	_, ok := err.(*stopWithRequeueDelay)
	return ok
}

func IsTerminal(err error) bool {
	return errors.Is(err, reconcile.TerminalError(nil))
}

func init() {
	StopAndForget = errors.New("stop and forget")
	StopWithRequeue = errors.New("stop with requeue")
}

type stopWithRequeueDelay struct {
	delay time.Duration
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
