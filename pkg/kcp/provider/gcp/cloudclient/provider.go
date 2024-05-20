package cloudclient

import (
	"context"
)

type SkrClientProvider[T any] func(ctx context.Context, saJsonKeyPath string) (T, error)
