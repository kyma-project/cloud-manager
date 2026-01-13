package client

import (
	"context"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		//TODO: implement Azure VPC Network client
		return nil, nil
	}
}
