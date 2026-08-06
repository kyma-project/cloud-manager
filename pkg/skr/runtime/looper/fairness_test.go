package looper

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
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
		require.False(t, l.processOne(0, col.cyclicQueue, "cyclic", l.cyclicReAdd, col.cyclicQueue.Add))
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

// TestNotificationGateConflictDrops: when a notification worker Gets an SKR that is already
// being served (gate held by the cyclic sleeve), it must DROP it — no re-add to either
// queue — because the in-flight connect already satisfies the notification intent.
func TestNotificationGateConflictDrops(t *testing.T) {
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))
	calls := 0
	l := newTestLooper(col, func(_ int, _ string) { calls++ })

	// SKR active; simulate the cyclic sleeve holding the gate.
	col.cyclicQueue.Add("k")
	col.notifQueue.Add("k")
	require.True(t, col.Gate().TryClaim("k"))

	// Notification worker processes "k" with production callbacks (reAdd = stamp connect,
	// onConflict = drop). It must hit the conflict, not call handle, and not re-queue.
	require.False(t, l.processOne(0, col.notifQueue, "notification", l.recordNotifConnect, func(string) {}))

	assert.Equal(t, 0, calls, "handle must not run on the notification conflict path")
	assert.Equal(t, 0, col.notifQueue.Len(), "notification conflict must drop (no notif re-add)")
	// Cyclic queue: only the original membership entry, nothing extra pushed by the conflict.
	assert.True(t, col.Contains("k"))

	col.Gate().Release("k")
	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestNotifyRateLimitCoalesces: two notifications for the same SKR within notifMinInterval
// of its last notification connect are coalesced — only the first enqueues; the second is
// dropped and counted. After the interval elapses, a notification enqueues again.
func TestNotifyRateLimitCoalesces(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	col := newTestCollection(fakeClock)
	col.notifMinInterval = 10 * time.Second

	col.cyclicQueue.Add("k") // must be an active member for Notify to accept

	// First notification enqueues (no prior connect stamp).
	col.Notify("k")
	assert.Equal(t, 1, col.notifQueue.Len(), "first notification must enqueue")

	// Simulate the notification connect completing (stamps last-connect = now).
	drainQueue(t, col.notifQueue, 1)
	col.recordNotifConnect("k")

	// A notification within the interval is coalesced (dropped).
	col.Notify("k")
	assert.Equal(t, 0, col.notifQueue.Len(), "notification within notifMinInterval must be dropped")

	// After the interval, a notification enqueues again.
	fakeClock.Step(11 * time.Second)
	col.Notify("k")
	assert.Equal(t, 1, col.notifQueue.Len(), "notification after notifMinInterval must enqueue")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// moduleActiveValue reads the current module-active gauge for a bare kymaName (all other
// labels empty, as the strand tests build label-free objects).
func moduleActiveValue(t *testing.T, kymaName string) float64 {
	t.Helper()
	return gaugeValue(t, metrics.SkrRuntimeModuleActiveCount.WithLabelValues(kymaName, "", "", "", "", ""))
}

// TestAddKymaOnActiveMemberIsNoOp is the deterministic regression for the AddKyma strand
// fix. The KCP Kyma reconciler periodically re-activates live SKRs via AddKyma. A
// re-activation of an already-active member must be a pure no-op on the cyclic queue:
//
//   - it must NOT re-enqueue the SKR (the unconditional pre-fix Add routed through the
//     workqueue's Touch, moving an already-queued member to the TAIL — the same reorder
//     class that starved the fairness tail, and the write that races an in-flight connect
//     and can strand the SKR), and
//   - it must NOT double-count the module-active gauge.
//
// The reorder is the directly observable pre/post-fix difference: pre-fix, re-activating a
// queued member Touches it to the tail (head changes); post-fix the queue order is
// untouched. The strand itself is a concurrent race (guarded by TestAddKymaConcurrentNoStrand
// and the looper_sim harness); this test pins the deterministic contract the fix rests on.
func TestAddKymaOnActiveMemberIsNoOp(t *testing.T) {
	ctx := context.Background()
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	const k = "kyma-addkyma-noop"
	before := moduleActiveValue(t, k)

	// First activation: enqueues once, counts once.
	col.AddKyma(ctx, kymaObj(k))
	require.True(t, col.Contains(k))
	require.Equal(t, 1, col.cyclicQueue.Len(), "first activation must enqueue exactly once")
	assert.InDelta(t, before+1, moduleActiveValue(t, k), 1e-9, "first activation must count module-active once")

	// Periodic re-activation of an already-active member: must be a no-op on the queue and
	// on the gauge.
	col.AddKyma(ctx, kymaObj(k))
	col.AddKyma(ctx, kymaObj(k))
	assert.Equal(t, 1, col.cyclicQueue.Len(),
		"re-activating an already-active SKR must NOT enqueue a duplicate")
	assert.InDelta(t, before+1, moduleActiveValue(t, k), 1e-9,
		"re-activating an already-active SKR must NOT double-count module-active")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestAddKymaDoesNotReorderQueuedMember pins the observable pre/post-fix difference. A
// queued cyclic member sits at a fixed rotation position. Pre-fix, AddKyma re-activation
// called l.cyclicQueue.Add unconditionally; for an already-queued item the workqueue calls
// slidingQueue.Touch, which moves it to the TAIL — reshuffling rotation order (the fairness
// hazard) and, concurrently with an in-flight connect, the write that can strand it. Post-fix
// the Add is guarded on !alreadyActive, so the re-activation leaves the queue untouched.
//
// Pre-fix this drains as [b, c, a] (a Touched to tail) → FAIL; post-fix [a, b, c].
func TestAddKymaDoesNotReorderQueuedMember(t *testing.T) {
	ctx := context.Background()
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	// Seed three active members in a fixed rotation order; "a" is at the head.
	col.AddKyma(ctx, kymaObj("a"))
	col.AddKyma(ctx, kymaObj("b"))
	col.AddKyma(ctx, kymaObj("c"))
	require.Equal(t, 3, col.cyclicQueue.Len())

	// Periodic re-activation of the head member. Pre-fix this Touches "a" to the tail.
	col.AddKyma(ctx, kymaObj("a"))

	require.Equal(t, 3, col.cyclicQueue.Len(), "re-activation must not change queue depth")
	got := make([]string, 0, 3)
	for range 3 {
		item, shutdown := col.cyclicQueue.Get()
		require.False(t, shutdown)
		got = append(got, item)
		col.cyclicQueue.Done(item)
	}
	assert.Equal(t, []string{"a", "b", "c"}, got,
		"re-activating an already-queued member must NOT reorder the cyclic rotation")

	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
}

// TestAddKymaConcurrentNoStrand is the concurrent regression for the strand: real cyclic
// workers rotate the fleet while a driver goroutine hammers AddKyma on already-active
// members — mirroring the KCP Kyma reconciler periodically re-activating live SKRs
// concurrently with the connect lifecycle (Get → handle → cyclicReAdd → Done). The invariant
// the fix guarantees: every active member remains serviceable — none is silently stranded
// (a member absent from the ready queue with nothing to re-add it). We assert every member is
// served repeatedly and none disappears from rotation over the run.
//
// Pre-fix, the unconditional external Add races the connect lifecycle and can leave a member
// out of both the dirty set and the ready queue → that member's service count flatlines.
// Post-fix, the external Add is a guarded no-op, so the cyclic worker's own success-path
// re-add is the sole writer and rotation stays complete.
func TestAddKymaConcurrentNoStrand(t *testing.T) {
	ctx := context.Background()
	col := newTestCollection(clocktesting.NewFakeClock(time.Now()))

	const fleet = 24
	const workers = 4
	items := make([]string, fleet)
	for i := range items {
		items[i] = fmt.Sprintf("kyma-%02d", i)
	}

	freq := newFreqType()
	freq.reset(items...)

	l := newTestLooper(col, func(_ int, kymaName string) { freq.inc(kymaName) })
	l.cyclicImmediateThreshold = 1 // large-fleet FIFO re-add (matches prod)

	for _, it := range items {
		col.cyclicQueue.Add(it)
	}

	// Real cyclic workers with the PRODUCTION re-add path.
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for !l.processOne(id, col.cyclicQueue, "cyclic", l.cyclicReAdd, col.cyclicQueue.Add) {
			}
		}(w)
	}

	// Driver: continuously re-activate already-active members (the KCP reconciler behavior
	// that exposed the strand). Goes through the REAL add() path where the race lives.
	var driver sync.WaitGroup
	driver.Add(1)
	go func() {
		defer driver.Done()
		for {
			select {
			case <-stop:
				return
			default:
				for _, it := range items {
					col.AddKyma(ctx, kymaObj(it))
				}
			}
		}
	}()

	// Let the rotation run long enough that every member should be served many times.
	assert.Eventually(t, func() bool {
		_, served, _, _ := freq.statsTracked()
		return served == fleet
	}, 5*time.Second, 10*time.Millisecond, "every active member must be served — none stranded")

	// Run a while longer so a strand (a member that stops cycling) would show as a stalled
	// service count while others climb.
	time.Sleep(500 * time.Millisecond)

	close(stop)
	driver.Wait()
	col.cyclicQueue.ShutDown()
	col.notifQueue.ShutDown()
	wg.Wait()

	// Fairness/liveness: every member served, and the spread stays bounded (a stranded
	// member would be served ~once while the rest climb into the hundreds → huge rel spread).
	cnt, served, _, rel := freq.statsTracked()
	assert.Equal(t, fleet, cnt)
	assert.Equal(t, fleet, served, "every active member must have been served at least once")
	assert.Less(t, rel, 1.0, "no member may be stranded — service spread must stay bounded")
}
