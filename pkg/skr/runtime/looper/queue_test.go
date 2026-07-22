package looper

import (
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/clock"
	clocktesting "k8s.io/utils/clock/testing"
)

type freqType struct {
	l  sync.Mutex
	mt map[string]int
	mu map[string]int
}

func newFreqType() *freqType {
	r := &freqType{
		l:  sync.Mutex{},
		mt: make(map[string]int),
		mu: make(map[string]int),
	}
	return r
}

func (f *freqType) reset(items ...string) {
	f.l.Lock()
	defer f.l.Unlock()
	f.mt = make(map[string]int)
	f.mu = make(map[string]int)
	for _, item := range items {
		f.mt[item] = 0
	}
}

func (f *freqType) inc(item string) {
	f.l.Lock()
	defer f.l.Unlock()
	if _, ok := f.mt[item]; ok {
		f.mt[item] += 1
	} else {
		f.mu[item] += 1
	}
}

func (f *freqType) unknownItems() map[string]int {
	f.l.Lock()
	defer f.l.Unlock()
	return f.mu
}

// statsTracked returns stats on tracked items
func (f *freqType) statsTracked() (int, int, int, float64) {
	f.l.Lock()
	defer f.l.Unlock()

	return f.stats(f.mt)
}

// stats returns:
// * item count
// * count of items with freqs greater than zero
// * difference between max and min freq
// * relative diff (diff/min)
func (f *freqType) stats(m map[string]int) (int, int, int, float64) {
	minVal := util.MaxInt
	maxVal := util.MinInt
	gz := 0
	for _, x := range m {
		if x > 0 {
			gz += 1
		}
		if x < minVal {
			minVal = x
		}
		if x > maxVal {
			maxVal = x
		}
	}

	d := maxVal - minVal
	rel := float64(0)
	if minVal != 0 {
		rel = float64(d) / float64(minVal)
	}
	return len(m), gz, d, rel
}

func (f *freqType) assertTracked(t *testing.T, expectedCount int) {
	cnt, gz, _, rel := f.statsTracked()
	assert.Equal(t, expectedCount, cnt)
	assert.Equal(t, expectedCount, gz)
	assert.Less(t, rel, 0.2)
}

func (f *freqType) assertUnknown(t *testing.T) {
	for k, v := range f.unknownItems() {
		if v > 1 {
			assert.Fail(t, "item %s unexpected more then one: %d", k, v)
		}
	}
}

// TestSmokeQueue ports the CyclicQueue smoke test onto the workqueue-backed Queue.
// It proves the cyclic-worker contract: the worker re-Adds a still-member item after
// Done so it cycles, load is balanced across workers, and Remove stops cycling
// (Done no longer re-adds — cycling is the worker's job now).
func TestSmokeQueue(t *testing.T) {
	first := []string{"a", "b", "c", "d", "e"}
	second := []string{"f", "g", "h", "i", "j"}
	var all []string
	all = append(all, first...)
	all = append(all, second...)

	q := NewQueue()
	freq := newFreqType()

	concurrency := 5
	for range concurrency {
		go func() {
			for {
				item, shutdown := q.Get()
				if shutdown {
					return
				}
				freq.inc(item)
				time.Sleep(100 * time.Millisecond)
				q.Done(item)
				// Cycling is the worker's job: re-Add while still a member.
				if q.Contains(item) {
					q.Add(item)
				}
			}
		}()
	}

	time.Sleep(1 * time.Second)
	freq.assertTracked(t, 0)
	freq.assertUnknown(t)

	for _, item := range all {
		q.Add(item)
	}
	freq.reset(all...)

	time.Sleep(3 * time.Second)

	freq.assertTracked(t, 10)
	freq.assertUnknown(t)

	for _, item := range first {
		q.Remove(item)
	}
	freq.reset(second...)

	time.Sleep(3 * time.Second)

	freq.assertTracked(t, 5)
	freq.assertUnknown(t)

	for _, item := range second {
		q.Remove(item)
	}
	freq.reset()

	time.Sleep(2 * time.Second)

	freq.assertTracked(t, 0)
	freq.assertUnknown(t)

	q.ShutDown()
}

// TestSlidingQueueTouchMovesToTail verifies the move-to-end behavior that Delay
// relies on: Touching a queued item pops it to the tail.
func TestSlidingQueueTouchMovesToTail(t *testing.T) {
	q := &slidingQueue[string]{}
	q.Push("a")
	q.Push("b")
	q.Push("c")

	q.Touch("a") // a -> tail, order becomes b, c, a

	assert.Equal(t, 3, q.Len())
	assert.Equal(t, "b", q.Pop())
	assert.Equal(t, "c", q.Pop())
	assert.Equal(t, "a", q.Pop())
	assert.Equal(t, 0, q.Len())
}

// TestSlidingQueueTouchMissing is a no-op when the item is not present.
func TestSlidingQueueTouchMissing(t *testing.T) {
	q := &slidingQueue[string]{}
	q.Push("a")
	q.Push("b")

	q.Touch("z")

	assert.Equal(t, 2, q.Len())
	assert.Equal(t, "a", q.Pop())
	assert.Equal(t, "b", q.Pop())
}

// TestQueueAddAfterHonorsDelay uses a fake clock to assert an item enqueued via
// AddAfter is not dispatchable until the delay elapses.
func TestQueueAddAfterHonorsDelay(t *testing.T) {
	fakeClock := clocktesting.NewFakeClock(time.Now())
	q := newQueueWithClock(fakeClock)

	base := fakeClock.Waiters() // waiter baseline BEFORE AddAfter schedules its timer
	q.AddAfter("a", 5*time.Second)

	// membership is recorded immediately even while the add is pending
	assert.True(t, q.Contains("a"))

	// nothing dispatchable yet
	assert.Eventually(t, func() bool { return q.Len() == 0 }, time.Second, 10*time.Millisecond)

	stepAfterWaiter(t, fakeClock, base, 5*time.Second)

	// the delayed add now lands in the queue
	assert.Eventually(t, func() bool { return q.Len() == 1 }, time.Second, 10*time.Millisecond)

	item, shutdown := q.Get()
	assert.False(t, shutdown)
	assert.Equal(t, "a", item)

	q.ShutDown()
}

// TestQueueDedup verifies the workqueue dirty-set dedup: adding the same item
// twice while queued yields a single dispatch.
func TestQueueDedup(t *testing.T) {
	q := newQueueWithClock(clock.RealClock{})

	q.Add("a")
	q.Add("a")

	assert.Equal(t, 1, q.Len())

	q.ShutDown()
}

// TestQueueRemoveStopsCycle verifies that after Remove the item is no longer a
// member, and Done never re-adds (cycling is the worker's job), so the item does
// not reappear in the queue.
func TestQueueRemoveStopsCycle(t *testing.T) {
	q := newQueueWithClock(clock.RealClock{})

	q.Add("a")
	item, shutdown := q.Get()
	assert.False(t, shutdown)
	assert.Equal(t, "a", item)

	q.Remove("a")
	assert.False(t, q.Contains("a"))

	q.Done("a") // Done never re-adds
	assert.Equal(t, 0, q.Len())

	q.ShutDown()
}

// TestQueueShutDownWithDrain lets in-flight workers finish before returning.
func TestQueueShutDownWithDrain(t *testing.T) {
	q := NewQueue()

	var wg sync.WaitGroup
	wg.Go(func() {
		for {
			item, shutdown := q.Get()
			if shutdown {
				return
			}
			// stop cycling this item so drain can complete
			q.Remove(item)
			q.Done(item)
		}
	})

	q.Add("a")
	q.ShutDownWithDrain()
	wg.Wait()
}
