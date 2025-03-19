package client

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type constantsSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *constantsSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *constantsSuite) TestParsePvVolumeHandle() {
	sampleVolumeHandle := "shoot--kyma-dev--c-6ea9b9b#f21d936aa5673444a95852a#pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca###default"
	resourceGroupName, storageAccountName, fileShareName, uuid, secretNamespace, err := ParsePvVolumeHandle(sampleVolumeHandle)
	suite.Nil(err)
	suite.Equal("shoot--kyma-dev--c-6ea9b9b", resourceGroupName)
	suite.Equal("f21d936aa5673444a95852a", storageAccountName)
	suite.Equal("pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca", fileShareName)
	suite.Equal("", uuid)
	suite.Equal("default", secretNamespace)
}

func (suite *constantsSuite) TestGetStorageAccountPath() {
	samplePath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testsa"
	suite.Equal(samplePath, GetStorageAccountPath("3f1d2fbd-117a-4742-8bde-6edbcdee6a04", "test-rg", "testsa"))
}

func (suite *constantsSuite) TestGetBackupPolicyPath() {

	// Arrange
	subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
	resourceGroupName := "kh-rg"
	vaultName := "kh-vault-service"
	backupPolicyName := "kh-backup-policy"

	expectedPath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/kh-rg/providers/Microsoft.RecoveryServices/vaults/kh-vault-service/backupPolicies/kh-backup-policy"

	// Act
	actualPath := GetBackupPolicyPath(subscriptionId, resourceGroupName, vaultName, backupPolicyName)

	// Assert
	suite.Equal(expectedPath, actualPath)

}

func (suite *constantsSuite) TestGetVaultPath() {

	// Arrange
	subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
	resourceGroupName := "kh-rg"
	vaultName := "kh-vault-service"

	expectedPath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/kh-rg/providers/Microsoft.RecoveryServices/vaults/kh-vault-service"

	// Act
	actualPath := GetVaultPath(subscriptionId, resourceGroupName, vaultName)

	// Assert
	suite.Equal(expectedPath, actualPath)

}

func (suite *constantsSuite) TestGetContainerName() {

	// Arrange
	resourceGroupName := "kh-rg"
	storageAccountName := "khstorageaccount"

	expectedPath := "StorageContainer;Storage;kh-rg;khstorageaccount"

	// Act
	actualPath := GetContainerName(resourceGroupName, storageAccountName)

	// Assert
	suite.Equal(expectedPath, actualPath)

}

func (suite *constantsSuite) TestGetRecoveryPointPath() {

	// Arrange
	subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
	resourceGroupName := "kh-rg"
	vaultName := "kh-vault-service"
	storageAccountName := "khstorageaccount"
	protectedItemName := "AzureFileShare;C269EB5A60C5955A69DAE32E9F5A1FDAE343AB5E8F0709DDE1B46E17D02E19DD"
	recoveryPointName := "966593861375688"

	expectedPath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/kh-rg/providers/Microsoft.RecoveryServices/vaults/kh-vault-service/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;kh-rg;khstorageaccount/protectedItems/AzureFileShare;C269EB5A60C5955A69DAE32E9F5A1FDAE343AB5E8F0709DDE1B46E17D02E19DD/recoveryPoints/966593861375688"

	// Act
	actualPath := GetRecoveryPointPath(subscriptionId, resourceGroupName, vaultName, storageAccountName, protectedItemName, recoveryPointName)

	// Assert
	suite.Equal(expectedPath, actualPath)

}

func TestConstants(t *testing.T) {
	suite.Run(t, new(constantsSuite))
}
