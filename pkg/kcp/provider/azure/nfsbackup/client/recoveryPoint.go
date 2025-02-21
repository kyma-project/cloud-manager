package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type RecoveryPointClient interface {
	GetRecoveryPoint(ctx context.Context)
	ListRecoveryPoints(ctx context.Context)
}

type recoveryPointClient struct {
	*armrecoveryservicesbackup.RecoveryPointsClient
}

func NewRecoveryPointClient(subscriptionId string, cred *azidentity.DefaultAzureCredential) (RecoveryPointClient, error) {

	rpc, err := armrecoveryservicesbackup.NewRecoveryPointsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return recoveryPointClient{rpc}, nil
}

func (c recoveryPointClient) GetRecoveryPoint(ctx context.Context) {
	// TODO: implementation details
}
func (c recoveryPointClient) ListRecoveryPoints(ctx context.Context) {
	// TODO: implementation details
}
