package client

import (
	"context"
)

type GardenClientProvider[T any] func(ctx context.Context, region, key, secret string) (T, error)

type SkrClientProvider[T any] func(ctx context.Context, region, key, secret, role string) (T, error)
