package sim

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/e2e/fake"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestManager(t *testing.T) {

	t.Run("start and stop with non-started cluster and two runnables without errors", func(t *testing.T) {
		clstr := fake.NewClusterBuilder().
			WithCache(false, true).
			Build()
		r1 := &fake.Runnable{}
		r2 := &fake.Runnable{}
		m := NewManager(clstr, logr.Discard())
		assert.NoError(t, m.Add(r1))
		assert.NoError(t, m.Add(r2))

		ctx, cancel := context.WithCancel(context.Background())

		var err error
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = m.Start(ctx)
		}()

		_ = wait.PollUntilContextTimeout(ctx, 2*time.Millisecond, 10*time.Millisecond, false, func(ctx context.Context) (bool, error) {
			return clstr.StartCallCount > 0 && r1.StartCallCount > 0 && r2.StartCallCount > 0, nil
		})

		assert.Equal(t, 1, clstr.StartCallCount)
		assert.Equal(t, 1, r1.StartCallCount)
		assert.Equal(t, 1, r2.StartCallCount)

		cancel()
		wg.Wait()

		assert.NoError(t, err)
	})

	t.Run("start and stop with already started cluster and two runnables without errors", func(t *testing.T) {
		clstr := fake.NewClusterBuilder().
			WithCache(true, true).
			Build()
		r1 := &fake.Runnable{}
		r2 := &fake.Runnable{}
		m := NewManager(clstr, logr.Discard())
		assert.NoError(t, m.Add(r1))
		assert.NoError(t, m.Add(r2))

		ctx, cancel := context.WithCancel(context.Background())

		var err error
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = m.Start(ctx)
		}()

		_ = wait.PollUntilContextTimeout(ctx, 2*time.Millisecond, 10*time.Millisecond, false, func(ctx context.Context) (bool, error) {
			return r1.StartCallCount > 0 && r2.StartCallCount > 0, nil
		})

		assert.Equal(t, 0, clstr.StartCallCount)
		assert.Equal(t, 1, r1.StartCallCount)
		assert.Equal(t, 1, r2.StartCallCount)

		cancel()
		wg.Wait()

		assert.NoError(t, err)
	})

	t.Run("start and stop with non-started cluster and two runnables with cluster error", func(t *testing.T) {
		clusterError := errors.New("cluster error")
		clstr := fake.NewClusterBuilder().
			WithCache(false, true).
			WithStartErrors(clusterError, nil).
			Build()
		r1 := &fake.Runnable{}
		r2 := &fake.Runnable{}
		m := NewManager(clstr, logr.Discard())
		assert.NoError(t, m.Add(r1))
		assert.NoError(t, m.Add(r2))

		ctx, cancel := context.WithCancel(context.Background())

		var err error
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = m.Start(ctx)
		}()

		cancel()
		wg.Wait()

		assert.True(t, errors.Is(err, clusterError))
	})

	t.Run("start and stop with non-started cluster and two runnables with all errors", func(t *testing.T) {
		err1 := errors.New("cluster error")
		err2 := errors.New("r1 error")
		err3 := errors.New("r2 error")
		clstr := fake.NewClusterBuilder().
			WithCache(false, true).
			WithStartErrors(nil, err1).
			Build()
		r1 := &fake.Runnable{ErrPost: err2}
		r2 := &fake.Runnable{ErrPost: err3}
		m := NewManager(clstr, logr.Discard())
		assert.NoError(t, m.Add(r1))
		assert.NoError(t, m.Add(r2))

		ctx, cancel := context.WithCancel(context.Background())

		var err error
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = m.Start(ctx)
		}()

		cancel()
		wg.Wait()

		assert.True(t, errors.Is(err, err1))
		assert.True(t, errors.Is(err, err2))
		assert.True(t, errors.Is(err, err3))
	})
}
