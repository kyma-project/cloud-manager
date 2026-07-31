package looper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clocktesting "k8s.io/utils/clock/testing"
)

// TestCyclicReAddFifoAboveThreshold: when the active fleet is at/above
// cyclicImmediateThreshold, the cyclic re-add uses an immediate FIFO Add — the SKR is
// dispatchable right away, with no AddAfter delay.
func TestCyclicReAddFifoAboveThreshold(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))
	l := newTestLooper(col, func(_ int, _ string) {})
	l.cyclicImmediateThreshold = 2

	// Two members → fleet >= threshold → FIFO path.
	col.cyclicQueue.Add("k1")
	col.cyclicQueue.Add("k2")
	// Drain both so nothing is queued before we exercise the re-add.
	drainQueue(t, col.cyclicQueue, 2)

	l.cyclicReAdd("k1")

	// Immediate Add → dispatchable now, no clock step needed.
	assert.Eventually(t, func() bool { return col.cyclicQueue.Len() == 1 }, time.Second, 10*time.Millisecond,
		"FIFO re-add must make the SKR dispatchable immediately")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestCyclicReAddFloorBelowThreshold: when the active fleet is below
// cyclicImmediateThreshold, the cyclic re-add uses AddAfter(cyclicMinInterval) — the SKR
// is NOT dispatchable until the interval elapses (the small-fleet hot-loop floor).
func TestCyclicReAddFloorBelowThreshold(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)
	l := newTestLooper(col, func(_ int, _ string) {})
	l.cyclicImmediateThreshold = 5

	// One member → fleet < threshold → AddAfter floor path.
	col.cyclicQueue.Add("k")
	drainQueue(t, col.cyclicQueue, 1)

	base := fakeClock.Waiters()
	l.cyclicReAdd("k")

	// Not dispatchable before the interval.
	assert.Never(t, func() bool { return col.cyclicQueue.Len() > 0 }, 200*time.Millisecond, 20*time.Millisecond,
		"below-threshold re-add must honor the cyclicMinInterval floor")

	// After the interval it reappears.
	stepAfterWaiter(t, fakeClock, base, 60*time.Second)
	assert.Eventually(t, func() bool { return col.cyclicQueue.Len() == 1 }, time.Second, 10*time.Millisecond)

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestCyclicReAddDropsNonMember: the re-add must not re-schedule a removed SKR (its own
// mid-flight "cleanup done" Remove), regardless of the threshold mode.
func TestCyclicReAddDropsNonMember(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))
	l := newTestLooper(col, func(_ int, _ string) {})
	l.cyclicImmediateThreshold = 1

	// Not a member (never added / already removed).
	l.cyclicReAdd("gone")
	assert.False(t, col.Contains("gone"), "re-add must not activate a non-member")
	assert.Equal(t, 0, col.cyclicQueue.Len())

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestCyclicFairDistribution is the regression for the unfair-distribution bug. With
// FIFO re-add (fleet >= threshold) the cyclic queue is a strict round-robin: an SKR
// re-added on Done goes to the tail, so a single worker draining the queue serves every
// SKR exactly once per pass, in the same order, every pass — regardless of how long any
// individual handle takes. After exactly K complete passes every SKR has been served
// exactly K times (rel spread == 0). Under the old AddAfter-into-readyAt-heap re-add,
// per-cycle completion-time variance random-walked each SKR's heap position and the
// spread was unbounded; this test would have caught that.
//
// A single worker is used deliberately: it makes the rotation order fully deterministic
// so the assertion has no dependence on goroutine scheduling (the earlier multi-worker
// version was flaky — stopping on a shared connect budget cut the final pass off at a
// scheduling-dependent point, leaving a large, nondeterministic spread). Handle-time
// variance is simulated per call but, by the FIFO invariant, must not affect fairness.
func TestCyclicFairDistribution(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	const fleet = 60
	const passes = 20
	items := make([]string, fleet)
	for i := range items {
		items[i] = fmt.Sprintf("kyma-%03d", i)
	}

	freq := newFreqType()
	freq.reset(items...)

	// Per-SKR handle "cost": a busy no-op whose iteration count varies by a per-key call
	// counter, simulating the 100ms-vs-10s completion-time variance that random-walked
	// each SKR's readyAt-heap position under the old AddAfter re-add. Under FIFO re-add
	// this variance must NOT affect rotation fairness.
	callCount := map[string]int{}
	handle := func(_ int, kymaName string) {
		n := callCount[kymaName]
		callCount[kymaName] = n + 1
		iters := 1000
		if n%2 == 1 {
			iters = 100000 // odd calls ~100x heavier
		}
		sink := 0
		for i := range iters {
			sink += i
		}
		_ = sink
		freq.inc(kymaName)
	}
	l := newTestLooper(col, handle)
	l.cyclicImmediateThreshold = 1 // FIFO mode for any non-empty fleet

	for _, it := range items {
		col.cyclicQueue.Add(it)
	}

	// Single worker, exactly fleet*passes guarded cycles → K complete FIFO passes, using
	// the PRODUCTION re-add (l.cyclicReAdd) so shipped logic is covered. Deterministic:
	// no concurrency, no clock steps (FIFO re-add is immediate Add, not AddAfter).
	for range fleet * passes {
		require.False(t, l.processOne(0, col.cyclicQueue, "cyclic", l.cyclicReAdd))
	}

	// Fairness: every SKR was served, and the spread is tight (in fact exactly `passes`).
	freq.assertTracked(t, fleet)

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestMembershipLen: MembershipLen counts active members (independent of queued depth),
// and Remove decrements it.
func TestMembershipLen(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	q := newQueueWithClock(fakeClock)

	assert.Equal(t, 0, q.MembershipLen())

	q.Add("a")
	q.Add("b")
	assert.Equal(t, 2, q.MembershipLen())

	// A delayed add is a member immediately even though it is not yet queued/ready:
	// MembershipLen counts it, Len (ready depth) does not.
	q.AddAfter("c", 60*time.Second)
	assert.Equal(t, 3, q.MembershipLen())
	assert.Less(t, q.Len(), 3, "delayed item is a member but not yet in the ready queue")

	q.Remove("a")
	assert.Equal(t, 2, q.MembershipLen())
	assert.False(t, q.Contains("a"))

	q.ShutDown()
}

// drainQueue Gets and Dones n items so the queue is empty before a test exercises a
// specific re-add, without leaving items in-flight (Done clears processing).
func drainQueue(t *testing.T, q *Queue, n int) {
	t.Helper()
	for range n {
		item, shutdown := q.Get()
		require.False(t, shutdown)
		q.Done(item)
	}
}
