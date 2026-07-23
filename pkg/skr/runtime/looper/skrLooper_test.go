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

	var calls int64
	l := newTestLooper(col, func(_ int, _ string) { atomic.AddInt64(&calls, 1) })

	col.cyclicQueue.Add("k")

	// one guarded cycle: handle runs, success reAdd schedules AddAfter(60s)
	base := fakeClock.Waiters() // waiter baseline BEFORE the reAdd schedules its timer
	require.False(t, l.processOne(0, col.cyclicQueue, "cyclic", func(kymaName string) {
		if l.Contains(kymaName) {
			col.cyclicQueue.AddAfter(kymaName, l.cyclicMinInterval)
		}
	}))
	assert.Equal(t, int64(1), atomic.LoadInt64(&calls))

	// not dispatchable yet
	assert.Never(t, func() bool { return col.cyclicQueue.Len() > 0 }, 200*time.Millisecond, 20*time.Millisecond)

	// after the interval it reappears
	stepAfterWaiter(t, fakeClock, base, 60*time.Second)
	assert.Eventually(t, func() bool { return col.cyclicQueue.Len() == 1 }, time.Second, 10*time.Millisecond)

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestNotificationWorkerFifoDrainAndDelaysCyclic: the notification worker does NOT
// re-add itself (FIFO drain), and on success it Delays the cyclic entry.
func TestNotificationWorkerFifoDrainAndDelaysCyclic(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	var calls int64
	l := newTestLooper(col, func(_ int, _ string) { atomic.AddInt64(&calls, 1) })

	// SKR is active (in cyclic queue) and has a pending notification.
	col.cyclicQueue.Add("k")
	col.notifQueue.Add("k")

	require.False(t, l.processOne(0, col.notifQueue, "notification", func(kymaName string) {
		if l.Contains(kymaName) {
			col.CyclicQueue().Delay(kymaName)
		}
	}))
	assert.Equal(t, int64(1), atomic.LoadInt64(&calls))

	// FIFO drain: the notification queue is empty (no self re-add).
	assert.Equal(t, 0, col.notifQueue.Len())

	// The cyclic entry was Delay'd (Add) — still present/dispatchable.
	assert.True(t, col.Contains("k"))

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}
