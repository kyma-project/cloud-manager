package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type ProtectionPoliciesClient interface {
	CreateBackupPolicy(ctx context.Context)
	DeleteBackupPolicy(ctx context.Context)
}

type protectionPoliciesClient struct {
	*armrecoveryservicesbackup.ProtectionPoliciesClient
}

func NewProtectionPoliciesClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (ProtectionPoliciesClient, error) {

	ppc, err := armrecoveryservicesbackup.NewProtectionPoliciesClient(subscriptionId, cred, nil)

	if err != nil {
		return nil, err
	}

	return protectionPoliciesClient{ppc}, nil
}

func (c protectionPoliciesClient) CreateBackupPolicy(ctx context.Context) {
	// TODO: implementation details
}

func (c protectionPoliciesClient) DeleteBackupPolicy(ctx context.Context) {
	// TODO: implementation details
}
