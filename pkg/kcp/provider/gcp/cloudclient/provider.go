package cloudclient

import (
	"context"
)

type ClientProvider[T any] func(ctx context.Context, saJsonKeyPath string) (T, error)
