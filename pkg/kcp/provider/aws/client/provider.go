package client

import (
	"context"
	"sync"
)

type GardenClientProvider[T any] func(ctx context.Context, region, key, secret string) (T, error)

type SkrClientProvider[T any] func(ctx context.Context, region, key, secret, role string) (T, error)

func NewCachedGardenClientProvider[T comparable](p GardenClientProvider[T]) GardenClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, region, key, secret string) (T, error) {
		m.Lock()
		defer m.Unlock()

		var err error
		if result == nilT {
			result, err = p(ctx, region, key, secret)
		}
		return result, err
	}
}

func NewCachedSkrClientProvider[T comparable](p SkrClientProvider[T]) SkrClientProvider[T] {
	var result T
	var nilT T
	var m sync.Mutex
	return func(ctx context.Context, region, key, secret, role string) (T, error) {
		m.Lock()
		defer m.Unlock()

		var err error
		if result == nilT {
			result, err = p(ctx, region, key, secret, role)
		}
		return result, err
	}
}
