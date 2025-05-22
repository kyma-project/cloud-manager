package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	CreateFileShare(ctx context.Context, id string) error
	GetFileShare(ctx context.Context, id string) (*armstorage.FileShareItem, error)
	DeleteFileShare(ctx context.Context, id string) error
}

type client struct {
	fileSharesClient *armstorage.FileSharesClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			return nil, err
		}

		fsc, err := NewFileShareClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		return fsc, nil
	}
}

func NewFileShareClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (Client, error) {

	fsc, err := armstorage.NewFileSharesClient(subscriptionId, cred, nil)

	if err != nil {
		return nil, err
	}

	return client{fsc}, nil
}

func (c client) CreateFileShare(ctx context.Context, id string) error {
	//This method is used for integration testing only, noop here
	//NOOP
	return nil
}

func (c client) GetFileShare(ctx context.Context, id string) (*armstorage.FileShareItem, error) {
	resourceGroupName, storageAccountName, fileShareName, _, _, err := azurerwxvolumebackupclient.ParsePvVolumeHandle(id)

	if err != nil {
		return nil, err
	}

	pager := c.fileSharesClient.NewListPager(resourceGroupName, storageAccountName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, fs := range page.Value {
			if ptr.Deref(fs.Name, "") == fileShareName {
				return fs, nil
			}
		}
	}

	return nil, nil
}

func (c client) DeleteFileShare(ctx context.Context, id string) error {
	resourceGroupName, storageAccountName, fileShareName, _, _, err := azurerwxvolumebackupclient.ParsePvVolumeHandle(id)
	if err != nil {
		return err
	}

	_, err = c.fileSharesClient.Delete(ctx, resourceGroupName, storageAccountName, fileShareName,
		&armstorage.FileSharesClientDeleteOptions{
			XMSSnapshot: nil,
			Include:     ptr.To("leased-snapshots"),
		})
	return err
}
