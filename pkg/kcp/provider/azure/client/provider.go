package client

import (
	"context"
)

type ClientProvider[T any] func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (T, error)
