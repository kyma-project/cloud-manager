package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRateLimiter(t *testing.T) {

	testWhen := func(r workqueue.TypedRateLimiter[client.Object], whenValues []time.Duration) {
		obj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "rate-limiter-obj",
			},
		}
		r.Forget(obj)
		for i, expected := range whenValues {
			actual := r.When(obj)
			//fmt.Printf("#%d %f  %f \n", i, expected.Seconds(), actual.Seconds())
			assert.Equal(t, expected, actual, "turn #%d", i)
		}
		r.Forget(obj)
		for i, expected := range whenValues {
			actual := r.When(obj)
			//fmt.Printf("#%d %f  %f \n", i, expected.Seconds(), actual.Seconds())
			assert.Equal(t, expected, actual, "turn 2x #%d", i)
		}
	}

	t.Run("Quick", func(t *testing.T) {
		testWhen(Quick, []time.Duration{
			100 * time.Millisecond,
			400 * time.Millisecond,
			1600 * time.Millisecond,
			6400 * time.Millisecond,
			25600 * time.Millisecond,
			102400 * time.Millisecond,
			409600 * time.Millisecond,
			10 * time.Minute,
			10 * time.Minute,
		})
	})

	t.Run("Slow1s", func(t *testing.T) {
		testWhen(Slow1s, []time.Duration{
			1 * time.Second,
			4 * time.Second,
			16 * time.Second,
			64 * time.Second,
			256 * time.Second,
			10 * time.Minute,
			10 * time.Minute,
			10 * time.Minute,
		})
	})

	t.Run("Slow10s", func(t *testing.T) {
		testWhen(Slow10s, []time.Duration{
			10 * time.Second,
			40 * time.Second,
			160 * time.Second,
			10 * time.Minute,
			10 * time.Minute,
			10 * time.Minute,
		})
	})

	t.Run("UltraSlow1m", func(t *testing.T) {
		testWhen(UltraSlow1m, []time.Duration{
			1 * time.Minute,
			2 * time.Minute,
			4 * time.Minute,
			8 * time.Minute,
			16 * time.Minute,
			32 * time.Minute,
			60 * time.Minute,
			60 * time.Minute,
			60 * time.Minute,
		})
	})

}
