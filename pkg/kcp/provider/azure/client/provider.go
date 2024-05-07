package client

import (
	"context"
)

type SkrClientProvider[T any] func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (T, error)
