package composed

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	ReconciliationLabelError        = "error"
	ReconciliationLabelRequeueAfter = "requeue_after"
	ReconciliationLabelRequeue      = "requeue"
	ReconciliationLabelSuccess      = "success"
	ReconciliationLabelCanceled     = "canceled"
	ReconciliationLabelDeadline     = "deadline"
)

func Handling() *Handler {
	return &Handler{}
}

type Handler struct {
	controller string
	name       string
	noLog      bool
	tracker    Tracker
}

func (h *Handler) WithNoLog() *Handler {
	h.noLog = true
	return h
}

func (h *Handler) WithMetrics(controller string, name string) *Handler {
	h.controller = controller
	h.name = name
	return h
}

func (h *Handler) WithName(name string) *Handler {
	h.name = name
	return h
}

func (h *Handler) WithTracker(t Tracker) *Handler {
	h.tracker = t
	return h
}

func (h *Handler) Handle(err error, ctx context.Context) (ctrl.Result, error) {
	var logger logr.Logger
	if h.noLog {
		logger = logr.Discard()
	} else {
		logger = LoggerFromCtx(ctx)
	}

	// used in defer func() to report the cloud_manager_reconcile metric
	result := ReconciliationLabelSuccess
	if h.controller != "" && h.name != "" {
		defer func() {
			Reconcile.WithLabelValues(h.controller, h.name, result).Inc()
		}()
	}
	if h.tracker != nil && h.name != "" {
		defer func() {
			h.tracker.Record(h.name, result)
		}()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		//logger.Info("Reconciliation finished with context deadline exceeded")
		result = ReconciliationLabelDeadline
		return ctrl.Result{}, nil
	}
	if errors.Is(err, context.Canceled) {
		//logger.Info("Reconciliation finished with context canceled")
		result = ReconciliationLabelCanceled
		return ctrl.Result{}, nil
	}
	if IsTerminal(err) {
		//logger.WithValues("err", err.Error()).Info("Reconciliation finished with terminal error")
		result = ReconciliationLabelError
		return ctrl.Result{}, err
	}
	if IsStopAndForget(err) {
		//logger.Info("Reconciliation finished with stop and forget")
		result = ReconciliationLabelSuccess
		return ctrl.Result{}, nil
	}
	if IsStopWithRequeue(err) {
		//logger.Info("Reconciliation finished with requeue")
		result = ReconciliationLabelRequeue
		return ctrl.Result{Requeue: true}, nil
	}
	if IsStopWithRequeueDelay(err) {
		var ed *stopWithRequeueDelay
		errors.As(err, &ed)
		//logger.WithValues("delay", ed.Delay().String()).Info("Reconciliation finished with requeue delayed")
		if ed.Delay() <= 0 {
			ed = &stopWithRequeueDelay{delay: time.Hour}
			logger.Info("Reconciliation requeue delayed set to 1h since it was zero")
		}
		result = ReconciliationLabelRequeueAfter
		return ctrl.Result{RequeueAfter: ed.Delay()}, nil
	}
	//logger.Info("Reconciliation finished without control error - doing stop and forget")
	result = ReconciliationLabelSuccess
	return ctrl.Result{}, err
}

func HandleWithoutLogging(err error, ctx context.Context) (ctrl.Result, error) {
	return Handling().WithNoLog().Handle(err, ctx)
}
