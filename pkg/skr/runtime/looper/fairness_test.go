package looper

import (
	"fmt"
	"sync"
	"sync/atomic"
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
// FIFO re-add (fleet >= threshold) and randomized-but-deterministic per-SKR handle
// durations (simulating the 100ms-vs-10s completion-time variance that random-walked
// each SKR's readyAt-heap position under the old AddAfter re-add), every SKR must be
// connected a near-equal number of times across many cycles. assertTracked checks the
// relative spread (max-min)/min < 0.2. Under the old AddAfter-into-readyAt-heap re-add
// this spread was unbounded; FIFO keeps it tight.
func TestCyclicFairDistribution(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	const fleet = 60
	items := make([]string, fleet)
	for i := range items {
		items[i] = fmt.Sprintf("kyma-%03d", i)
	}

	freq := newFreqType()
	freq.reset(items...)

	// Per-SKR deterministic handle "cost": a busy no-op whose iteration count varies
	// widely by a per-key call counter, simulating the 100ms-vs-10s completion-time
	// variance that random-walked each SKR's readyAt-heap position under the old AddAfter
	// re-add. No wall-clock sleeps — deterministic and fast. Under FIFO re-add this
	// variance must NOT affect rotation fairness.
	var mu sync.Mutex
	callCount := map[string]int{}
	handle := func(_ int, kymaName string) {
		mu.Lock()
		n := callCount[kymaName]
		callCount[kymaName] = n + 1
		mu.Unlock()
		// vary the work: even calls cheap, odd calls ~100x heavier
		iters := 1000
		if n%2 == 1 {
			iters = 100000
		}
		sink := 0
		for i := 0; i < iters; i++ {
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

	// Run a fixed number of guarded cycles across several workers, using the PRODUCTION
	// re-add (l.cyclicReAdd) so the test covers shipped logic. Stop after enough total
	// connects that each SKR should have been served many times in a fair rotation.
	const workers = 8
	const totalConnects = fleet * 20
	var remaining int64 = totalConnects
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := range workers {
		go func(id int) {
			defer wg.Done()
			for atomic.AddInt64(&remaining, -1) >= 0 {
				if l.processOne(id, col.cyclicQueue, "cyclic", l.cyclicReAdd) {
					return // shutting down
				}
			}
		}(w)
	}
	wg.Wait()

	// Fairness: every tracked SKR was served, and the spread is tight.
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
