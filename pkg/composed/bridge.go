package composed

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func HandleWithoutLogging(err error, ctx context.Context) (ctrl.Result, error) {
	logger := logr.Discard()
	ctx = LoggerIntoCtx(ctx, logger)
	return Handle(err, ctx)
}

func Handle(err error, ctx context.Context) (ctrl.Result, error) {
	logger := LoggerFromCtx(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Info("Reconciliation finished with context deadline exceeded")
		return ctrl.Result{}, nil
	}
	if errors.Is(err, context.Canceled) {
		logger.Info("Reconciliation finished with context canceled")
		return ctrl.Result{}, nil
	}
	if IsTerminal(err) {
		logger.WithValues("err", err.Error()).Info("Reconciliation finished with terminal error")
		return ctrl.Result{}, err
	}
	if IsStopAndForget(err) {
		logger.Info("Reconciliation finished with stop and forget")
		return ctrl.Result{}, nil
	}
	if IsStopWithRequeue(err) {
		logger.Info("Reconciliation finished with requeue")
		return ctrl.Result{Requeue: true}, nil
	}
	if IsStopWithRequeueDelay(err) {
		var ed *stopWithRequeueDelay
		errors.As(err, &ed)
		logger.WithValues("delay", ed.Delay().String()).Info("Reconciliation finished with requeue delayed")
		return ctrl.Result{RequeueAfter: ed.Delay()}, nil
	}
	logger.Info("Reconciliation finished without control error - doing stop and forget")
	return ctrl.Result{}, err
}
