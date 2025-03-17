package client

import (
	"context"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

func NewMockClient() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, _, _, _, _ string, _ ...string) (Client, error) {
		jobsMock = &jobsMockClient{
			jobsClient: *newJobsMockClient(),
		}
		restoreMock = &restoreMockClient{
			restoreClient: *newRestoreMockClient(),
		}
		return client{
			nil,
			nil,
			nil,
			nil,
			jobsMock,
			restoreMock,
			nil,
			nil,
		}, nil
	}
}
