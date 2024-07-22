package looper

import (
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type freqType struct {
	l  sync.Mutex
	mt map[interface{}]int
	mu map[interface{}]int
}

func newFreqType() *freqType {
	r := &freqType{
		l:  sync.Mutex{},
		mt: make(map[interface{}]int),
		mu: make(map[interface{}]int),
	}
	return r
}

func (f *freqType) reset(items ...interface{}) {
	f.l.Lock()
	defer f.l.Unlock()
	f.mt = make(map[interface{}]int)
	f.mu = make(map[interface{}]int)
	for _, item := range items {
		f.mt[item] = 0
	}
}

func (f *freqType) inc(item interface{}) {
	f.l.Lock()
	defer f.l.Unlock()
	if _, ok := f.mt[item]; ok {
		f.mt[item] += 1
	} else {
		f.mu[item] += 1
	}
}

func (f *freqType) trackedItems() map[interface{}]int {
	f.l.Lock()
	defer f.l.Unlock()
	return f.mt
}

func (f *freqType) unknownItems() map[interface{}]int {
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

// statsUnknown returns stats on tracked items
func (f *freqType) statsUnknown() (int, int, int, float64) {
	f.l.Lock()
	defer f.l.Unlock()

	return f.stats(f.mu)
}

// stats returns:
// * item count
// * count of items with freqs greater than zero
// * difference between max and min freq
// * relative diff (diff/min)
func (f *freqType) stats(m map[interface{}]int) (int, int, int, float64) {
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

func (f *freqType) print() {
	for v, ff := range f.mt {
		fmt.Printf("%v: %v\n", v, ff)
	}
	for v, ff := range f.mu {
		fmt.Printf("%v: %v\n", v, ff)
	}
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

func TestSmokeCyclicQueue(t *testing.T) {
	first := []interface{}{"a", "b", "c", "d", "e"}
	second := []interface{}{"f", "g", "h", "i", "j"}
	var all []interface{}
	all = append(all, first...)
	all = append(all, second...)

	q := NewCyclicQueue()
	freq := newFreqType()

	concurrency := 5
	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				itemX, shutdown := q.Get()
				if shutdown {
					return
				}
				item := itemX.(string)
				freq.inc(item)
				time.Sleep(100 * time.Millisecond)
				q.Done(item)
			}
		}()
	}

	time.Sleep(1 * time.Second)
	freq.assertTracked(t, 0)
	freq.assertUnknown(t)

	q.Add(all...)
	freq.reset(all...)

	time.Sleep(3 * time.Second)

	freq.assertTracked(t, 10)
	freq.assertUnknown(t)

	q.Remove(first...)
	freq.reset(second...)

	time.Sleep(3 * time.Second)

	freq.assertTracked(t, 5)
	freq.assertUnknown(t)

	q.Remove(second...)
	freq.reset()

	time.Sleep(2 * time.Second)

	freq.assertTracked(t, 0)
	freq.assertUnknown(t)

	q.Shutdown()

}
