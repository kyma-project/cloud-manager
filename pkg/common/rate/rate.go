package rate

import (
	"fmt"
	"math"
	"sync"
	"time"

	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// Quick to be used instead of `Requeue: true` with base value of 0.1s and cap of 10min (600s)
	//	#                         controller tests
	//	0     0.1s                0.01s
	//	1     0.4s                0.04s
	//	2     1.6s                0.16s
	//	3     6.4s                0.64s
	//	4     25.6s               1s (2.56)
	//	5     102.4s
	//	6     409.6s
	//	7     600s (1638.4)
	Quick workqueue.TypedRateLimiter[client.Object]

	// Slow1s with base value of 1s and cap of 10min (600s)
	//	#                        controller tests
	//	0     1s                 0.1s
	//	1     4s                 0.4s
	//	2     16s                1s (1.6)
	//	3     64s
	//	4     256s
	//	5     600s (1024)
	//	6     409.6s
	//	7     600s (1638.4)
	Slow1s workqueue.TypedRateLimiter[client.Object]

	// Slow10s with base value of 10s and cap of 10min (600s)
	//	#                        controller tests
	//	0     10s                 0.1s
	//	1     40s                 0.4s
	//	2    160s                 1s (1.6)
	//	3    600s (640)
	Slow10s workqueue.TypedRateLimiter[client.Object]

	// UltraSlow1m with base value of 1m and cap of 1h
	//	#                        controller tests
	//	0    1m                 0.1s
	//	1    2m                 0.4s
	//	2    4m                 1s (1.6)
	//	3    8m
	//  4    16m
	//  5    32m
	//  6    60m (64)
	UltraSlow1m workqueue.TypedRateLimiter[client.Object]
)

func init() {
	Quick = NewObjectRateLimiter(
		NewItemExponentialFailureRateLimiter(100*time.Millisecond, 10*time.Minute, 2),
	)

	Slow1s = NewObjectRateLimiter(
		NewItemExponentialFailureRateLimiter(1*time.Second, 10*time.Minute, 2),
	)

	Slow10s = NewObjectRateLimiter(
		NewItemExponentialFailureRateLimiter(10*time.Second, 10*time.Minute, 2),
	)

	UltraSlow1m = NewObjectRateLimiter(
		NewItemExponentialFailureRateLimiter(1*time.Minute, 1*time.Hour, 1),
	)

}

// ObjectRateLimiter

func NewObjectRateLimiter(inner workqueue.TypedRateLimiter[string]) workqueue.TypedRateLimiter[client.Object] {
	return &ObjectRateLimiter{
		inner: inner,
	}
}

type ObjectRateLimiter struct {
	inner workqueue.TypedRateLimiter[string]
}

func (r *ObjectRateLimiter) When(item client.Object) time.Duration {
	return r.inner.When(r.objectKey(item))
}

func (r *ObjectRateLimiter) NumRequeues(item client.Object) int {
	return r.inner.NumRequeues(r.objectKey(item))
}

func (r *ObjectRateLimiter) Forget(item client.Object) {
	r.inner.Forget(r.objectKey(item))
}

func (r *ObjectRateLimiter) objectKey(obj client.Object) string {
	var gk *schema.GroupKind
	if obj.GetObjectKind().GroupVersionKind().Group != "" && obj.GetObjectKind().GroupVersionKind().Kind != "" {
		gk = ptr.To(obj.GetObjectKind().GroupVersionKind().GroupKind())
	} else if arr, _, err := commonscheme.KcpScheme.ObjectKinds(obj); err == nil {
		gk = ptr.To(arr[0].GroupKind())
	} else if arr, _, err := commonscheme.SkrScheme.ObjectKinds(obj); err == nil {
		gk = ptr.To(arr[0].GroupKind())
	} else if arr, _, err := commonscheme.GardenScheme.ObjectKinds(obj); err == nil {
		gk = ptr.To(arr[0].GroupKind())
	} else {
		gk = ptr.To(schema.GroupKind{Kind: fmt.Sprintf("%T", obj)})
	}

	key := fmt.Sprintf("%s/%s/%s", gk.String(), obj.GetNamespace(), obj.GetName())

	return key
}

// ItemExponentialExpiringFailureRateLimiter =================================================

type ItemExponentialExpiringFailureRateLimiter struct {
	failuresLock sync.Mutex
	failures     *cache.Cache

	baseDelay time.Duration
	maxDelay  time.Duration
	speed     float64
}

var _ workqueue.TypedRateLimiter[string] = &ItemExponentialExpiringFailureRateLimiter{}

func NewItemExponentialFailureRateLimiter(baseDelay time.Duration, maxDelay time.Duration, speed float64) workqueue.TypedRateLimiter[string] {
	expiration := time.Duration(int64(float64(maxDelay) * 1.3))
	cleanup := time.Duration(int64(float64(maxDelay) * 1.4))
	return &ItemExponentialExpiringFailureRateLimiter{
		failures:  cache.New(expiration, cleanup),
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		speed:     speed,
	}
}

func (r *ItemExponentialExpiringFailureRateLimiter) When(item string) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	exp := 0
	v, found := r.failures.Get(item)
	if found {
		exp = v.(int)
	}
	r.failures.Set(item, exp+1, cache.DefaultExpiration)

	// The backoff is capped such that 'calculated' value never overflows.
	backoff := float64(r.baseDelay.Nanoseconds()) * math.Pow(2, r.speed*float64(exp))
	if backoff > math.MaxInt64 {
		return r.maxDelay
	}

	calculated := time.Duration(backoff)
	if calculated > r.maxDelay {
		return r.maxDelay
	}

	return calculated
}

func (r *ItemExponentialExpiringFailureRateLimiter) NumRequeues(item string) int {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	exp := 0
	v, found := r.failures.Get(item)
	if found {
		exp = v.(int)
	}

	return exp
}

func (r *ItemExponentialExpiringFailureRateLimiter) Forget(item string) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	r.failures.Delete(item)
}
