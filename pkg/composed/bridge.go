package composed

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Handle(err error, _ context.Context) (ctrl.Result, error) {
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
		ed := err.(*stopWithRequeueDelay)
		return ctrl.Result{RequeueAfter: ed.Delay()}, nil
	}
	return ctrl.Result{}, err
}
