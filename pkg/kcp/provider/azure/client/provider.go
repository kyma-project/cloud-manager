package client

import (
	"context"
)

type ClientProvider[T any] func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (T, error)
