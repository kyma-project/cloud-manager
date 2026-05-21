package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

type StorageAccountClient interface {
	CreateStorageAccount(ctx context.Context, resourceGroupName string, accountName string, parameters armstorage.AccountCreateParameters, options *armstorage.AccountsClientBeginCreateOptions) (Poller[armstorage.AccountsClientCreateResponse], error)
	GetStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (armstorage.AccountsClientGetPropertiesResponse, error)
	ListStorageAccounts(ctx context.Context) ([]*armstorage.Account, error)
	DeleteStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientDeleteOptions) (armstorage.AccountsClientDeleteResponse, error)
}

func NewStorageAccountClient(svcAccounts *armstorage.AccountsClient) StorageAccountClient {
	return &storageAccountClient{svcAccounts: svcAccounts}
}

type storageAccountClient struct {
	svcAccounts *armstorage.AccountsClient
}

func (c *storageAccountClient) CreateStorageAccount(ctx context.Context, resourceGroupName string, accountName string, parameters armstorage.AccountCreateParameters, options *armstorage.AccountsClientBeginCreateOptions) (Poller[armstorage.AccountsClientCreateResponse], error) {
	return c.svcAccounts.BeginCreate(ctx, resourceGroupName, accountName, parameters, options)
}

func (c *storageAccountClient) GetStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (armstorage.AccountsClientGetPropertiesResponse, error) {
	return c.svcAccounts.GetProperties(ctx, resourceGroupName, accountName, options)
}

func (c *storageAccountClient) ListStorageAccounts(ctx context.Context) ([]*armstorage.Account, error) {
	pager := c.svcAccounts.NewListPager(&armstorage.AccountsClientListOptions{})

	var items []*armstorage.Account

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, resp.Value...)
	}

	return items, nil
}

func (c *storageAccountClient) DeleteStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientDeleteOptions) (armstorage.AccountsClientDeleteResponse, error) {
	return c.svcAccounts.Delete(ctx, resourceGroupName, accountName, options)
}
