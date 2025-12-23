package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type RecoveryPointClient interface {
	GetRecoveryPoint(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string, recoveryPointId string) (*armrecoveryservicesbackup.RecoveryPointResource, error)
	ListRecoveryPoints(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string) ([]*armrecoveryservicesbackup.RecoveryPointResource, error)
}

type recoveryPointClient struct {
	azureClient *armrecoveryservicesbackup.RecoveryPointsClient
}

func NewRecoveryPointClient(rpc *armrecoveryservicesbackup.RecoveryPointsClient) RecoveryPointClient {
	return recoveryPointClient{rpc}
}

func (c recoveryPointClient) GetRecoveryPoint(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string, recoveryPointId string) (*armrecoveryservicesbackup.RecoveryPointResource, error) {
	resp, err := c.azureClient.Get(
		ctx,
		vaultName,
		resourceGroupName,
		fabricName,
		containerName,
		protectedItemName,
		recoveryPointId,
		nil,
	)

	if err != nil {
		return nil, err
	}

	result := &armrecoveryservicesbackup.RecoveryPointResource{
		ETag:       resp.ETag,
		Location:   resp.Location,
		Properties: resp.Properties,
		Tags:       resp.Tags,
		ID:         resp.ID,
		Name:       resp.Name,
		Type:       resp.Type,
	}

	return result, nil
}
func (c recoveryPointClient) ListRecoveryPoints(ctx context.Context, vaultName string, resourceGroupName string, fabricName string, containerName string, protectedItemName string) ([]*armrecoveryservicesbackup.RecoveryPointResource, error) {
	var result []*armrecoveryservicesbackup.RecoveryPointResource
	pager := c.azureClient.NewListPager(
		vaultName,
		resourceGroupName,
		fabricName,
		containerName,
		protectedItemName,
		nil,
	)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return result, err
		}

		result = append(result, page.Value...)

	}

	return result, nil
}
