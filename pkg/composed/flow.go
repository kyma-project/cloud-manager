package composed

import (
	"context"
	"time"
)

func StopAndForgetAction(_ context.Context, _ State) (error, context.Context) {
	return StopAndForget, nil
}

func StopWithRequeueAction(_ context.Context, _ State) (error, context.Context) {
	return StopWithRequeue, nil
}

func StopWithRequeueDelayAction(d time.Duration) Action {
	return func(_ context.Context, _ State) (error, context.Context) {
		return StopWithRequeueDelay(d), nil
	}
}
