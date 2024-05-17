package composed

import (
	"context"
	"errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Handle(err error, _ context.Context) (ctrl.Result, error) {
	if errors.Is(err, context.DeadlineExceeded) {
		return ctrl.Result{}, nil
	}
	if IsTerminal(err) {
		return ctrl.Result{}, err
	}
	if IsStopAndForget(err) {
		return ctrl.Result{}, nil
	}
	if IsStopWithRequeue(err) {
		return ctrl.Result{Requeue: true}, nil
	}
	if IsStopWithRequeueDelay(err) {
		var ed *stopWithRequeueDelay
		errors.As(err, &ed)
		return ctrl.Result{RequeueAfter: ed.Delay()}, nil
	}
	return ctrl.Result{}, err
}
