package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/elliotchance/pie/v2"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func newStorageAccountStore(subscription string) *storageAccountStore {
	return &storageAccountStore{
		subscription: subscription,
	}
}

type storageAccountStore struct {
	m            sync.Mutex
	subscription string

	storageAccounts []*armstorage.Account
}

var _ azureclient.StorageAccountClient = (*storageAccountStore)(nil)

func (s *storageAccountStore) CreateStorageAccount(ctx context.Context, resourceGroupName string, accountName string, parameters armstorage.AccountCreateParameters, options *armstorage.AccountsClientBeginCreateOptions) (azureclient.Poller[armstorage.AccountsClientCreateResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, storageAccount := range s.storageAccounts {
		if ptr.Deref(storageAccount.Name, "") == accountName {
			return nil, fmt.Errorf("storage account %s already exists", accountName)
		}
	}

	id := azureutil.NewStorageAccountResourceId(s.subscription, resourceGroupName, accountName)

	props := &armstorage.AccountProperties{}
	if parameters.Properties != nil {
		if err := util.JsonCloneInto(parameters.Properties, props); err != nil {
			return nil, err
		}
	}
	props.ProvisioningState = new(armstorage.ProvisioningStateSucceeded)

	sa := &armstorage.Account{
		ID:         new(id.String()),
		Name:       new(accountName),
		Type:       new("Microsoft.Storage/storageAccounts"),
		Kind:       parameters.Kind,
		Location:   parameters.Location,
		SKU:        util.Must(util.Clone(parameters.SKU)),
		Properties: props,
		Tags:       parameters.Tags,
	}
	s.storageAccounts = append(s.storageAccounts, sa)

	return NewPollerMock(armstorage.AccountsClientCreateResponse{
		Account: *util.Must(util.Clone(sa)),
	}, nil, ""), nil
}

func (s *storageAccountStore) GetStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (armstorage.AccountsClientGetPropertiesResponse, error) {
	var result armstorage.AccountsClientGetPropertiesResponse
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, storageAccount := range s.storageAccounts {
		if ptr.Deref(storageAccount.Name, "") == accountName {
			result.Account = *util.Must(util.Clone(storageAccount))
			return result, nil
		}
	}

	return result, azuremeta.NewAzureNotFoundError()
}

func (s *storageAccountStore) ListStorageAccounts(ctx context.Context) ([]*armstorage.Account, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var result []*armstorage.Account
	for _, storageAccount := range s.storageAccounts {
		result = append(result, util.Must(util.Clone(storageAccount)))
	}

	return result, nil
}

func (s *storageAccountStore) DeleteStorageAccount(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientDeleteOptions) (armstorage.AccountsClientDeleteResponse, error) {
	var result armstorage.AccountsClientDeleteResponse
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	found := false
	for _, storageAccount := range s.storageAccounts {
		if ptr.Deref(storageAccount.Name, "") == accountName {
			found = true
			break
		}
	}
	if !found {
		return result, azuremeta.NewAzureNotFoundError()
	}

	s.storageAccounts = pie.FilterNot(s.storageAccounts, func(sa *armstorage.Account) bool {
		return ptr.Deref(sa.Name, "") == accountName
	})

	return result, nil
}
