package looper

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
)

// slidingQueue is a slice-backed workqueue.Queue whose Touch moves an already
// queued item to the tail. workqueue calls Touch when Add is invoked for an item
// that is already queued and not yet being processed; this gives us the
// move-to-end ("Delay") behavior the two-sleeve design relies on. All methods are
// called by workqueue from a single goroutine while holding its lock, so no
// locking is needed here.
type slidingQueue[T comparable] []T

func (q *slidingQueue[T]) Push(item T) {
	*q = append(*q, item)
}

func (q *slidingQueue[T]) Len() int {
	return len(*q)
}

func (q *slidingQueue[T]) Pop() (item T) {
	item = (*q)[0]
	// allow gc from the underlying array
	(*q)[0] = *new(T)
	*q = (*q)[1:]
	return item
}

// Touch moves item to the tail of the queue. If item is not present (should not
// happen, workqueue only calls Touch for queued items) it is a no-op.
func (q *slidingQueue[T]) Touch(item T) {
	s := *q
	for i, x := range s {
		if x == item {
			copy(s[i:], s[i+1:])
			s[len(s)-1] = *new(T)
			*q = append(s[:len(s)-1], item)
			return
		}
	}
}

// Queue is the SKR looper work queue. It wraps a workqueue delaying queue (which
// provides Add/AddAfter/Get/Done/ShutDown plus dirty-set dedup and free metrics)
// and adds a membership set to back Contains/Remove/Items, which the workqueue
// itself does not provide.
//
// Cycling is NOT a queue property: the cyclic worker re-schedules itself via
// AddAfter(CyclicMinInterval) on the success path, and the notification worker
// drains FIFO without re-adding. Done therefore just marks processing complete.
type Queue struct {
	wq workqueue.TypedDelayingInterface[string]

	mu         sync.Mutex
	membership map[string]struct{}
}

func NewQueue() *Queue {
	return newQueueWithClock(clock.RealClock{})
}

func newQueueWithClock(c clock.WithTicker) *Queue {
	return &Queue{
		wq: workqueue.NewTypedDelayingQueueWithConfig(workqueue.TypedDelayingQueueConfig[string]{
			Clock: c,
			Queue: workqueue.NewTypedWithConfig(workqueue.TypedQueueConfig[string]{
				Clock: c,
				Queue: &slidingQueue[string]{},
			}),
		}),
		membership: map[string]struct{}{},
	}
}

func (q *Queue) Add(item string) {
	q.mu.Lock()
	q.membership[item] = struct{}{}
	q.mu.Unlock()
	q.wq.Add(item)
}

// AddAfter enqueues item after the given delay. Membership is recorded now so a
// concurrent Contains/Remove sees the pending item.
func (q *Queue) AddAfter(item string, d time.Duration) {
	q.mu.Lock()
	q.membership[item] = struct{}{}
	q.mu.Unlock()
	q.wq.AddAfter(item, d)
}

// Delay re-adds an already queued item; via slidingQueue.Touch this moves it to
// the tail. If the item is not queued, Add enqueues it normally.
func (q *Queue) Delay(item string) {
	q.Add(item)
}

func (q *Queue) Get() (item string, shutdown bool) {
	return q.wq.Get()
}

// Done marks processing complete. Cycling is driven by the cyclic worker's
// explicit AddAfter, not here, so Done never re-adds.
func (q *Queue) Done(item string) {
	q.wq.Done(item)
}

func (q *Queue) Len() int {
	return q.wq.Len()
}

func (q *Queue) Contains(item string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	_, ok := q.membership[item]
	return ok
}

// Remove clears queue membership only. It does not abort in-flight processing;
// the next Done for the item will not re-add it because it is no longer a member.
func (q *Queue) Remove(item string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.membership, item)
}

func (q *Queue) Items() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	items := make([]string, 0, len(q.membership))
	for item := range q.membership {
		items = append(items, item)
	}
	return items
}

func (q *Queue) ShutDown() {
	q.wq.ShutDown()
}

func (q *Queue) ShutDownWithDrain() {
	q.wq.ShutDownWithDrain()
}

func (q *Queue) ShuttingDown() bool {
	return q.wq.ShuttingDown()
}
