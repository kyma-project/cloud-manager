package looper

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/clock"
	clocktesting "k8s.io/utils/clock/testing"
)

// newTestCollection builds an activeSkrCollection whose queues use the given clock,
// so tests can drive AddAfter delays deterministically. The same clock backs the
// notification rate limiter, and a small default notifMinInterval is set so tests that
// don't care about rate limiting are unaffected (they use distinct keys / single notifies).
func newTestCollection(c clock.WithTicker) *activeSkrCollection {
	return &activeSkrCollection{
		cyclicQueue:      newQueueWithClock(c),
		notifQueue:       newQueueWithClock(c),
		gate:             NewSkrGate(),
		clock:            c,
		notifMinInterval: 0, // disabled by default in tests; rate-limit tests set it explicitly
		lastNotifConnect: map[string]time.Time{},
	}
}

// newTestLooper wires a skrLooper over col with a stub handler and small delays. It
// does NOT set managerFactory/registry — handleFn replaces handleOneSkr so no live
// per-SKR manager is created. cyclicImmediateThreshold defaults to 1 (FIFO re-add for
// any non-empty fleet); tests needing the AddAfter small-fleet path pass their own
// reAdd closure to processOne, or set the field explicitly.
func newTestLooper(col *activeSkrCollection, handle func(id int, kymaName string)) *skrLooper {
	return &skrLooper{
		ActiveSkrCollectionAdmin: col,
		handleFn:                 handle,
		cyclicMinInterval:        60 * time.Second,
		reconcileTimeout:         10 * time.Second,
		cyclicImmediateThreshold: 1,
	}
}

// stepAfterWaiter waits until a new fake-clock waiter appears beyond base — the
// waiter count captured *before* the AddAfter we intend to fire — then steps the
// clock. This closes the race where workqueue's waitingLoop registers its NewTimer
// asynchronously after AddAfter returns: stepping before that timer is registered
// leaves it stuck with a past targetTime until the 10s heartbeat, well past the
// assert.Eventually window.
func stepAfterWaiter(t *testing.T, clk *clocktesting.FakeClock, base int, d time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool { return clk.Waiters() > base },
		2*time.Second, 5*time.Millisecond,
		"AddAfter timer was never registered on the fake clock")
	clk.Step(d)
}

func counterValue(t *testing.T, c prometheus.Counter) float64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, c.Write(m))
	return m.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	m := &dto.Metric{}
	require.NoError(t, g.Write(m))
	return m.GetGauge().GetValue()
}

// TestSkrGateAtomicity: N goroutines racing TryClaim on one key → exactly one wins;
// after Release a fresh TryClaim succeeds; Release is idempotent.
func TestSkrGateAtomicity(t *testing.T) {
	g := NewSkrGate()

	const n = 64
	var wins int64
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			if g.TryClaim("k") {
				atomic.AddInt64(&wins, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(1), atomic.LoadInt64(&wins), "exactly one goroutine must win the claim")
	assert.True(t, g.heldForTest("k"))

	// A second attempt while held fails.
	assert.False(t, g.TryClaim("k"))

	// After release, a fresh claim succeeds.
	g.Release("k")
	assert.False(t, g.heldForTest("k"))
	assert.True(t, g.TryClaim("k"))

	// Release is idempotent — releasing twice, or a never-held key, is safe.
	g.Release("k")
	g.Release("k")
	g.Release("never-held")
	assert.False(t, g.heldForTest("k"))
}

// TestWorkerGateConflict: with a stub that claims-and-blocks, a second worker Getting
// the same key does NOT call the stub, increments the conflict counter, and re-adds via
// a plain FIFO tail append (NOT AddAfter — that would reshuffle the cyclic order and is
// the fairness bug this design avoids). The re-added item is dispatchable immediately;
// once the first worker releases the claim, the retry wins.
func TestWorkerGateConflict(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	release := make(chan struct{})
	var calls int64
	handle := func(_ int, kymaName string) {
		atomic.AddInt64(&calls, 1)
		<-release // hold the claim until the test lets go
	}
	l := newTestLooper(col, handle)

	// Activate the SKR and pre-claim it as if worker A were mid-flight.
	col.cyclicQueue.Add("k")
	require.True(t, col.Gate().TryClaim("k"))

	before := counterValue(t, metrics.SkrLooperGateConflictTotal.WithLabelValues("cyclic"))

	// Worker B processes "k": it must lose the claim, bump the conflict counter, and
	// re-add via plain FIFO (q.Add). processOne blocks on Get, so run it in a goroutine
	// but it should return quickly (conflict path does not call handle).
	col.cyclicQueue.Add("k") // dirty-set dedup: still one dispatchable "k"
	done := make(chan bool, 1)
	go func() {
		done <- l.processOne(0, col.cyclicQueue, "cyclic", func(string) {}, col.cyclicQueue.Add)
	}()

	select {
	case shuttingDown := <-done:
		assert.False(t, shuttingDown)
	case <-time.After(2 * time.Second):
		t.Fatal("processOne did not return on the conflict path")
	}

	assert.Equal(t, int64(0), atomic.LoadInt64(&calls), "handle must NOT be called on the conflict path")
	after := counterValue(t, metrics.SkrLooperGateConflictTotal.WithLabelValues("cyclic"))
	assert.Equal(t, before+1, after, "conflict counter must increment")

	// The conflict requeue is a plain FIFO tail add (no delay): "k" is dispatchable again
	// immediately, without any clock step.
	assert.Eventually(t, func() bool { return col.cyclicQueue.Len() == 1 }, time.Second, 10*time.Millisecond,
		"conflict re-add must be a plain FIFO tail append, dispatchable immediately")

	// Release worker A's claim; worker B retries and now wins the claim → handle runs.
	col.Gate().Release("k")
	go func() {
		l.processOne(0, col.cyclicQueue, "cyclic", func(string) {}, col.cyclicQueue.Add)
	}()
	assert.Eventually(t, func() bool { return atomic.LoadInt64(&calls) == 1 }, 2*time.Second, 10*time.Millisecond)

	close(release)
	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestWorkerRemovalMidFlight: a blocking stub holds a claim; Remove(key) must NOT
// abort it nor free the gate claim; after release the claim clears and the key is
// not re-added (it is no longer a member).
func TestWorkerRemovalMidFlight(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	entered := make(chan struct{})
	release := make(chan struct{})
	handle := func(_ int, _ string) {
		close(entered)
		<-release
	}
	l := newTestLooper(col, handle)

	col.cyclicQueue.Add("k")

	go func() {
		l.processOne(0, col.cyclicQueue, "cyclic", func(kymaName string) {
			// mirrors cyclicWorker: guard the re-add on membership
			if l.Contains(kymaName) {
				col.cyclicQueue.AddAfter(kymaName, l.cyclicMinInterval)
			}
		}, col.cyclicQueue.Add)
	}()

	<-entered // worker is inside handle, holding the claim
	assert.True(t, col.Gate().heldForTest("k"))

	// Removal mid-flight: clears membership only.
	col.cyclicQueue.Remove("k")
	assert.False(t, col.Contains("k"))
	assert.True(t, col.Gate().heldForTest("k"), "Remove must NOT free the gate claim")

	// Let the manager finish; deferred Release frees the claim.
	close(release)
	assert.Eventually(t, func() bool { return !col.Gate().heldForTest("k") }, 2*time.Second, 10*time.Millisecond)

	// Because Remove cleared membership before the success-path reAdd ran, the guarded
	// reAdd must NOT re-schedule the removed SKR. Nothing lands in the cyclic queue.
	fakeClock.Step(60 * time.Second)
	assert.Never(t, func() bool { return col.cyclicQueue.Len() > 0 }, 300*time.Millisecond, 30*time.Millisecond)
	assert.False(t, col.Contains("k"), "removed SKR must not be re-activated by its own final cyclic pass")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestWorkerMembershipGuard (Guard 1): a queued key is Removed then delivered via a
// stale AddAfter; the worker drops it — no TryClaim, no handle, no re-add.
func TestWorkerMembershipGuard(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)

	var calls int64
	l := newTestLooper(col, func(_ int, _ string) { atomic.AddInt64(&calls, 1) })

	// Deliver "k" but it is NOT a member (never activated).
	col.cyclicQueue.wq.Add("k") // bypass membership recording to simulate a truly stale item
	require.Equal(t, 1, col.cyclicQueue.Len())

	shuttingDown := l.processOne(0, col.cyclicQueue, "cyclic", func(string) {
		t.Fatal("reAdd must not run for a dropped item")
	}, func(string) {
		t.Fatal("onConflict must not run for a dropped item")
	})
	assert.False(t, shuttingDown)
	assert.Equal(t, int64(0), atomic.LoadInt64(&calls), "handle must not be called for a non-member")
	assert.False(t, col.Gate().heldForTest("k"), "no claim must be taken for a non-member")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}
