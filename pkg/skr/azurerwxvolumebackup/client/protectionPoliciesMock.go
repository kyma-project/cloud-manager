package client

import (
	"context"
	"errors"
)

func newProtectionPoliciesMockClient() *protectionPoliciesClient {
	return &protectionPoliciesClient{}
}

type protectionPoliciesMockClient struct {
	protectionPoliciesClient
}

func (m *protectionPoliciesMockClient) CreateBackupPolicy(ctx context.Context, vaultName, resourceGroupName, policyName string) error {

	if vaultName == "vaultName - fail CreateBackupPolicy" {
		return errors.New("failed to create backup polic")
	}

	return nil
}

func (m *protectionPoliciesMockClient) DeleteBackupPolicy(ctx context.Context, vaultName, resourceGroupName, policyName string) error {

	// TODO: implement

	return nil
}
