package v2

import (
	"context"
	"time"
)

func StopAndForgetAction(ctx context.Context) (context.Context, error) {
	return ctx, StopAndForget
}

func StopWithRequeueAction(ctx context.Context) (context.Context, error) {
	return ctx, StopWithRequeue
}

func StopWithRequeueDelayAction(d time.Duration) Action {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, StopWithRequeueDelay(d)
	}
}
