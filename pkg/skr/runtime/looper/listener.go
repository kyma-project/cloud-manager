package looper

import (
	"context"

	"github.com/go-logr/logr"
	watcherevent "github.com/kyma-project/runtime-watcher/listener/pkg/v2/event"
	watchertypes "github.com/kyma-project/runtime-watcher/listener/pkg/v2/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

// NotificationComponentName is the runtime-watcher component name for cloud-manager.
// It becomes the /v2/{component}/event path on KCP and MUST match the `manager:`
// field of the Watcher CR (see config/watcher/watcher.yaml).
const NotificationComponentName = "cloud-manager"

// notificationListener is a manager.Runnable that owns the runtime-watcher
// SKREventListener HTTP server and an adapter goroutine that forwards each
// received notification to the notification sleeve via notify (== Notify).
//
// runtime-id == kymaName, so no lookup layer is needed: the runtime-id carried
// by the notification is the SKR's kymaName directly.
type notificationListener struct {
	listener *watcherevent.SKREventListener
	notify   func(kymaName string)
	logger   logr.Logger
}

var _ manager.Runnable = (*notificationListener)(nil)

// NewNotificationListener constructs the runtime-watcher listener bound to addr
// under the given componentName, forwarding valid notifications to notify.
func NewNotificationListener(addr, componentName string, notify func(kymaName string), logger logr.Logger) *notificationListener {
	l := watcherevent.NewSKREventListener(addr, componentName)
	l.Logger = logger.WithName("skr-notification-listener")
	return &notificationListener{
		listener: l,
		notify:   notify,
		logger:   l.Logger,
	}
}

// Start launches the adapter goroutine and runs the listener's HTTP server.
// It blocks until ctx is done; the adapter goroutine exits when the listener
// closes its events channel on shutdown.
func (n *notificationListener) Start(ctx context.Context) error {
	go n.adapt(n.listener.ReceivedEvents())
	return n.listener.Start(ctx)
}

// adapt drains received notifications, extracts the runtime-id (== kymaName),
// and forwards it to notify. Notifications with a missing/invalid runtime-id are
// dropped. Returns when ch is closed.
func (n *notificationListener) adapt(ch <-chan watchertypes.GenericEvent) {
	for evt := range ch {
		kymaName, ok := runtimeIDFromEvent(evt)
		if !ok {
			metrics.SkrLooperNotificationDroppedTotal.Inc()
			n.logger.V(1).Info("dropping SKR notification with missing/invalid runtime-id")
			continue
		}
		metrics.SkrLooperNotificationReceivedTotal.Inc()
		n.notify(kymaName)
	}
}

// runtimeIDFromEvent extracts the SKR runtime-id (== kymaName) from a
// runtime-watcher GenericEvent. Returns ("", false) if the payload is
// missing/malformed so the caller can drop it safely.
func runtimeIDFromEvent(evt watchertypes.GenericEvent) (string, bool) {
	if evt.Object == nil {
		return "", false
	}
	v, ok := evt.Object.Object["runtime-id"]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}
