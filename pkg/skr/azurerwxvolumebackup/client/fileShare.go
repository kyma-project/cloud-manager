package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type FileShareClient interface {
	CreateFileShare(ctx context.Context, id string) error
	GetFileShare(ctx context.Context, id string) (*armstorage.FileShareItem, error)
	DeleteFileShare(ctx context.Context, id string) error
}

type fileShareClient struct {
	azureFsClient *armstorage.FileSharesClient
}

func NewFileShareClientProvider() azureclient.ClientProvider[FileShareClient] {
	return FileShareClientProvider(NewClientProvider())
}

func FileShareClientProvider(backupClientProvider azureclient.ClientProvider[Client]) azureclient.ClientProvider[FileShareClient] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (FileShareClient, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		fsc, err := armstorage.NewFileSharesClient(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		return &fileShareClient{azureFsClient: fsc}, nil
	}
}

func (c *fileShareClient) CreateFileShare(ctx context.Context, id string) error {
	//This method is used for integration testing only, noop here
	//NOOP
	return nil
}

func (c *fileShareClient) GetFileShare(ctx context.Context, id string) (*armstorage.FileShareItem, error) {
	resourceGroupName, storageAccountName, fileShareName, _, _, err := ParsePvVolumeHandle(id)

	if err != nil {
		return nil, err
	}

	pager := c.azureFsClient.NewListPager(resourceGroupName, storageAccountName, nil)
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

func (c *fileShareClient) DeleteFileShare(ctx context.Context, id string) error {
	resourceGroupName, storageAccountName, fileShareName, _, _, err := ParsePvVolumeHandle(id)
	if err != nil {
		return err
	}

	_, err = c.azureFsClient.Delete(ctx, resourceGroupName, storageAccountName, fileShareName,
		&armstorage.FileSharesClientDeleteOptions{
			XMSSnapshot: nil,
			Include:     ptr.To("leased-snapshots"),
		})
	return err
}
