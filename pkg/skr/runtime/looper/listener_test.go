package looper

import (
	"context"
	"testing"
	"time"

	watchertypes "github.com/kyma-project/runtime-watcher/listener/pkg/v2/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clocktesting "k8s.io/utils/clock/testing"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

// eventWithRuntimeID builds a runtime-watcher GenericEvent carrying the given
// runtime-id value at Object["runtime-id"].
func eventWithRuntimeID(v any) watchertypes.GenericEvent {
	return watchertypes.GenericEvent{
		Object: &unstructured.Unstructured{Object: map[string]any{"runtime-id": v}},
	}
}

func TestRuntimeIDFromEvent(t *testing.T) {
	tests := []struct {
		name   string
		evt    watchertypes.GenericEvent
		want   string
		wantOk bool
	}{
		{"valid", eventWithRuntimeID("kyma-1"), "kyma-1", true},
		{"empty string", eventWithRuntimeID(""), "", false},
		{"non-string value", eventWithRuntimeID(42), "", false},
		{"missing key", watchertypes.GenericEvent{Object: &unstructured.Unstructured{Object: map[string]any{"other": "x"}}}, "", false},
		{"nil object", watchertypes.GenericEvent{Object: nil}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := runtimeIDFromEvent(tt.evt)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNotificationListenerAdapt: the adapter forwards valid runtime-ids to notify
// and bumps the received counter; malformed events are dropped and bump the
// dropped counter; the goroutine exits when the channel closes.
func TestNotificationListenerAdapt(t *testing.T) {
	received0 := counterValue(t, metrics.SkrLooperNotificationReceivedTotal)
	dropped0 := counterValue(t, metrics.SkrLooperNotificationDroppedTotal)

	var got []string
	n := &notificationListener{
		notify: func(kymaName string) { got = append(got, kymaName) },
	}

	ch := make(chan watchertypes.GenericEvent)
	done := make(chan struct{})
	go func() {
		n.adapt(ch)
		close(done)
	}()

	ch <- eventWithRuntimeID("kyma-a")
	ch <- eventWithRuntimeID("") // dropped: empty
	ch <- eventWithRuntimeID(42) // dropped: non-string
	ch <- eventWithRuntimeID("kyma-b")
	close(ch)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("adapt goroutine did not exit after channel close")
	}

	assert.Equal(t, []string{"kyma-a", "kyma-b"}, got, "only valid runtime-ids forwarded, in order")
	assert.Equal(t, float64(2), counterValue(t, metrics.SkrLooperNotificationReceivedTotal)-received0)
	assert.Equal(t, float64(2), counterValue(t, metrics.SkrLooperNotificationDroppedTotal)-dropped0)
}

// TestNotificationListenerAdaptToRealCollection: wiring the adapter into a real
// activeSkrCollection composes with Phase-3 Notify drop semantics — a notification
// for an inactive SKR enqueues nothing; after activation it enqueues into the
// notification queue.
func TestNotificationListenerAdaptToRealCollection(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))
	n := &notificationListener{notify: col.Notify}

	ch := make(chan watchertypes.GenericEvent)
	done := make(chan struct{})
	go func() {
		n.adapt(ch)
		close(done)
	}()

	// The adapter loop is single-threaded: receive -> notify -> next receive. Because
	// ch is unbuffered, a send that returns proves the PREVIOUS event's notify already
	// completed (the adapter was back at receive). barrier() exploits that to observe
	// notify's effect deterministically without sleeping.
	barrier := func() { ch <- eventWithRuntimeID("__barrier__") }

	// Inactive SKR: Notify is a no-op. "__barrier__" is also inactive → also dropped.
	ch <- eventWithRuntimeID("k1")
	barrier()
	assert.Equal(t, 0, col.NotificationQueue().Len(), "notification for inactive SKR must be dropped")

	// Activate, then notify: enqueues into the notification queue.
	col.AddKyma(context.Background(), kymaObj("k1"))
	ch <- eventWithRuntimeID("k1")
	barrier()
	assert.Equal(t, 1, col.NotificationQueue().Len(), "notification for active SKR must enqueue")

	close(ch)
	<-done
}
