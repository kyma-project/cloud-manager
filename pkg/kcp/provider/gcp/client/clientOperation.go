package client

import (
	"context"

	"github.com/googleapis/gax-go/v2"
)

type Operation[T any] interface {
	Name() string
	Done() bool
	Wait(ctx context.Context, opts ...gax.CallOption) (T, error)
}

type VoidOperation interface {
	Name() string
	Done() bool
	Wait(ctx context.Context, opts ...gax.CallOption) error
}
