package util

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextErrors(t *testing.T) {
	
	t.Run("IgnoreContextCanceled", func(t *testing.T) {
		
		t.Run("negative", func(t *testing.T) {
			err := errors.New("some error")
			err = IgnoreContextCanceled(err)
			assert.Error(t, err)
		})

		t.Run("simple", func(t *testing.T) {
			err := context.Canceled
			err = IgnoreContextCanceled(err)
			assert.NoError(t, err)
		})

		t.Run("real context", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			<-ctx.Done()
			err := IgnoreContextCanceled(ctx.Err())
			assert.NoError(t, err)
		})
	})

	t.Run("IgnoreContextDeadlineExceeded", func(t *testing.T) {

		t.Run("negative", func(t *testing.T) {
			err := errors.New("some error")
			err = IgnoreContextDeadlineExceeded(err)
			assert.Error(t, err)
		})

		t.Run("simple", func(t *testing.T) {
			err := context.DeadlineExceeded
			err = IgnoreContextDeadlineExceeded(err)
			assert.NoError(t, err)
		})

	})

}
