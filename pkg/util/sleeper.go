package util

import (
	"context"
	"time"
)

type Sleeper interface {
	Sleep(ctx context.Context, d time.Duration)
}

type SleeperFunc func(context.Context, time.Duration)

func (f SleeperFunc) Sleep(ctx context.Context, d time.Duration) {
	f(ctx, d)
}

var _ Sleeper = (SleeperFunc)(nil)

func RealSleeperFunc(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

var _ SleeperFunc = RealSleeperFunc

var _ Sleeper = SleeperFunc(RealSleeperFunc)
