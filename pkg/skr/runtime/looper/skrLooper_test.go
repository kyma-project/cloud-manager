package looper

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clocktesting "k8s.io/utils/clock/testing"
)

func kymaObj(name string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetName(name)
	return u
}

func scopeObj(name string) *cloudcontrolv1beta1.Scope {
	s := &cloudcontrolv1beta1.Scope{}
	s.SetName(name)
	return s
}

// TestActiveSkrCollectionRouting: AddKyma/AddScope populate the cyclic queue;
// Notify populates the notification queue only for active SKRs.
func TestActiveSkrCollectionRouting(t *testing.T) {
	ctx := context.Background()
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	// Activation goes to the cyclic queue.
	col.AddKyma(ctx, kymaObj("k1"))
	col.AddScope(ctx, scopeObj("k2"))
	assert.True(t, col.Contains("k1"))
	assert.True(t, col.Contains("k2"))
	assert.Equal(t, 2, col.CyclicQueue().Len())
	assert.Equal(t, 0, col.NotificationQueue().Len(), "activation must not touch the notification queue")

	// Notify for an active SKR enqueues into the notification queue.
	col.Notify("k1")
	assert.Equal(t, 1, col.NotificationQueue().Len())

	// Notify for an inactive/unknown SKR is dropped.
	col.Notify("unknown")
	assert.Equal(t, 1, col.NotificationQueue().Len(), "notification for an inactive SKR must be dropped")

	// Notify after removal is dropped.
	col.RemoveKyma(ctx, kymaObj("k1"))
	assert.False(t, col.Contains("k1"))
	col.Notify("k1")
	// still just the one queued earlier; membership drop prevents a new enqueue
	assert.LessOrEqual(t, col.NotificationQueue().Len(), 1)

	col.CyclicQueue().ShutDown()
	col.NotificationQueue().ShutDown()
}

// TestNotifyDropAfterRemove: Remove clears both queues' membership; a subsequent
// Notify enqueues nothing.
func TestNotifyDropAfterRemove(t *testing.T) {
	ctx := context.Background()
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	col.AddKyma(ctx, kymaObj("k"))
	col.RemoveKyma(ctx, kymaObj("k"))

	assert.False(t, col.Contains("k"))
	col.Notify("k")
	assert.Equal(t, 0, col.NotificationQueue().Len())

	col.CyclicQueue().ShutDown()
	col.NotificationQueue().ShutDown()
}

// TestCyclicWorkerReschedules: the cyclic worker re-adds the SKR after
// cyclicMinInterval on the success path; it is not dispatchable before the interval.
func TestCyclicWorkerReschedules(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	var calls atomic.Int64
	l := newTestLooper(col, func(_ int, _ string) { calls.Add(1) })

	col.cyclicQueue.Add("k")

	// one guarded cycle: handle runs, success reAdd schedules AddAfter(60s)
	base := fakeClock.Waiters() // waiter baseline BEFORE the reAdd schedules its timer
	require.False(t, l.processOne(0, col.cyclicQueue, "cyclic", func(kymaName string) {
		if l.Contains(kymaName) {
			col.cyclicQueue.AddAfter(kymaName, l.cyclicMinInterval)
		}
	}, col.cyclicQueue.Add))
	assert.Equal(t, int64(1), calls.Load())

	// not dispatchable yet
	assert.Never(t, func() bool { return col.cyclicQueue.Len() > 0 }, 200*time.Millisecond, 20*time.Millisecond)

	// after the interval it reappears
	stepAfterWaiter(t, fakeClock, base, 60*time.Second)
	assert.Eventually(t, func() bool { return col.cyclicQueue.Len() == 1 }, time.Second, 10*time.Millisecond)

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestNotificationWorkerFifoDrainDoesNotTouchCyclic: the notification worker drains its
// own queue FIFO (no self re-add) and MUST NOT write into the cyclic queue. Re-adding
// into cyclic here (the old CyclicQueue().Delay) moved an already-queued SKR to the tail
// via the workqueue's Touch, letting hot SKRs leapfrog the rotation and starve the tail.
func TestNotificationWorkerFifoDrainDoesNotTouchCyclic(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	var calls atomic.Int64
	l := newTestLooper(col, func(_ int, _ string) { calls.Add(1) })

	// Seed the cyclic queue with a fixed order; "k" sits behind others waiting its turn.
	col.cyclicQueue.Add("a")
	col.cyclicQueue.Add("k")
	col.cyclicQueue.Add("b")
	col.notifQueue.Add("k")

	// Process the notification for "k" with the PRODUCTION notification callbacks:
	// reAdd is a no-op and onConflict drops — neither touches the cyclic queue.
	require.False(t, l.processOne(0, col.notifQueue, "notification", func(string) {}, func(string) {}))
	assert.Equal(t, int64(1), calls.Load())

	// FIFO drain: the notification queue is empty (no self re-add).
	assert.Equal(t, 0, col.notifQueue.Len())

	// The cyclic queue is untouched: "k" kept its position (was NOT moved to the tail),
	// so the head is still "a". Drain and assert the original order a, k, b.
	require.Equal(t, 3, col.cyclicQueue.Len(), "notification must not add/remove cyclic entries")
	got := make([]string, 0, 3)
	for range 3 {
		item, shutdown := col.cyclicQueue.Get()
		require.False(t, shutdown)
		got = append(got, item)
		col.cyclicQueue.Done(item)
	}
	assert.Equal(t, []string{"a", "k", "b"}, got, "notification must not reorder the cyclic queue")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}
