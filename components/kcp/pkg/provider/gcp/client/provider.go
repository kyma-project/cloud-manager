package client

import (
	"context"
	"sync"
)


type ClientProvider[T any] func(ctx context.Context, saJsonKeyPath string) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, saJsonKeyPath string) (T, error) {
		m.Lock()
		defer m.Unlock()

		var err error
		if result == nilT {
			result, err = p(ctx, saJsonKeyPath)
		}
		return result, err
	}
}
