package looper

import (
	"github.com/elliotchance/pie/v2"
	"sync"
)

type empty struct{}

// type t interface{}
type set map[interface{}]empty

func (s set) has(item interface{}) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item interface{}) {
	s[item] = empty{}
}

func (s set) delete(item interface{}) {
	delete(s, item)
}

func (s set) len() int {
	return len(s)
}

func NewCyclicQueue() *CyclicQueue {
	return &CyclicQueue{
		cond:       sync.NewCond(&sync.Mutex{}),
		all:        set{},
		processing: set{},
	}
}

type CyclicQueue struct {
	cond         *sync.Cond
	shuttingDown bool
	queue        []interface{}
	all          set
	processing   set
}

func (q *CyclicQueue) Items() []interface{} {
	arr := make([]interface{}, q.Len())
	copy(arr, q.queue)
	return arr
}

func (q *CyclicQueue) Contains(item interface{}) bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	_, ok := q.all[item]
	return ok
}

func (q *CyclicQueue) Add(items ...interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	for _, item := range items {
		if q.all.has(item) {
			return
		}

		q.all.insert(item)
		q.queue = append(q.queue, item)
	}

	q.cond.Signal()
}

func (q *CyclicQueue) Remove(items ...interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	for _, item := range items {
		if q.all.has(item) {
			q.all.delete(item)
			q.queue = pie.Filter(q.queue, func(x interface{}) bool {
				return x != item
			})
		}
	}
}

func (q *CyclicQueue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

func (q *CyclicQueue) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return nil, true
	}
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}

	item = q.queue[0]
	// allow gc from the underlying array
	q.queue[0] = nil
	q.queue = q.queue[1:]

	q.processing.insert(item)

	return item, false
}

func (q *CyclicQueue) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.delete(item)
	if q.all.has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

func (q *CyclicQueue) Shutdown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.shuttingDown = true
	q.cond.Broadcast()
}

func (q *CyclicQueue) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}
