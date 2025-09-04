package client

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type constantsSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *constantsSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *constantsSuite) TestParsePvVolumeHandle() {
	sampleVolumeHandle := "shoot--kyma-dev--c-6ea9b9b#f21d936aa5673444a95852a#pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca###default"
	resourceGroupName, storageAccountName, fileShareName, uuid, secretNamespace, err := ParsePvVolumeHandle(sampleVolumeHandle)
	s.Nil(err)
	s.Equal("shoot--kyma-dev--c-6ea9b9b", resourceGroupName)
	s.Equal("f21d936aa5673444a95852a", storageAccountName)
	s.Equal("pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca", fileShareName)
	s.Equal("", uuid)
	s.Equal("default", secretNamespace)
}

func (s *constantsSuite) TestGetStorageAccountPath() {
	samplePath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testsa"
	s.Equal(samplePath, GetStorageAccountPath("3f1d2fbd-117a-4742-8bde-6edbcdee6a04", "test-rg", "testsa"))
}

func (s *constantsSuite) TestGetBackupPolicyPath() {

	// Arrange
	subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
	resourceGroupName := "kh-rg"
	vaultName := "kh-vault-service"
	backupPolicyName := "kh-backup-policy"

	expectedPath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/kh-rg/providers/Microsoft.RecoveryServices/vaults/kh-vault-service/backupPolicies/kh-backup-policy"

	// Act
	actualPath := GetBackupPolicyPath(subscriptionId, resourceGroupName, vaultName, backupPolicyName)

	// Assert
	s.Equal(expectedPath, actualPath)

}

func (s *constantsSuite) TestGetVaultPath() {

	// Arrange
	subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a04"
	resourceGroupName := "kh-rg"
	vaultName := "kh-vault-service"

	expectedPath := "/subscriptions/3f1d2fbd-117a-4742-8bde-6edbcdee6a04/resourceGroups/kh-rg/providers/Microsoft.RecoveryServices/vaults/kh-vault-service"

	// Act
	actualPath := GetVaultPath(subscriptionId, resourceGroupName, vaultName)

	// Assert
	s.Equal(expectedPath, actualPath)

}

func (s *constantsSuite) TestGetContainerName() {

	// Arrange
	resourceGroupName := "kh-rg"
	storageAccountName := "khstorageaccount"

	expectedPath := "StorageContainer;Storage;kh-rg;khstorageaccount"

	// Act
	actualPath := GetContainerName(resourceGroupName, storageAccountName)

	// Assert
	s.Equal(expectedPath, actualPath)

}

func (s *constantsSuite) TestGetRecoveryPointPath() {

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
	s.Equal(expectedPath, actualPath)

}

func (s *constantsSuite) TestAzureStorageErrorInfoToJson() {
	var errorCode int32 = 1234
	var sampleErrors = make([]*armrecoveryservicesbackup.AzureStorageErrorInfo, 3)
	var recommendations = make([]*string, 3)
	recommendations[0] = to.Ptr("recommendation 1")
	recommendations[1] = to.Ptr("recommendation 2")
	recommendations[2] = to.Ptr("recommendation 3")
	sampleErrors[0] = &armrecoveryservicesbackup.AzureStorageErrorInfo{
		ErrorCode:       to.Ptr(errorCode),
		ErrorString:     to.Ptr("Sample message 1"),
		Recommendations: recommendations,
	}
	sampleErrors[1] = &armrecoveryservicesbackup.AzureStorageErrorInfo{
		ErrorCode:       to.Ptr(errorCode),
		ErrorString:     to.Ptr("Sample message 2"),
		Recommendations: recommendations,
	}
	sampleErrors[2] = &armrecoveryservicesbackup.AzureStorageErrorInfo{
		ErrorCode:       to.Ptr(errorCode),
		ErrorString:     to.Ptr("Sample message 3"),
		Recommendations: recommendations,
	}
	detailsInJson, err := AzureStorageErrorInfoToJson(sampleErrors)
	s.Nil(err)
	s.Equal("[{\"errorCode\":1234,\"errorString\":\"Sample message 1\",\"recommendations\":[\"recommendation 1\",\"recommendation 2\",\"recommendation 3\"]},{\"errorCode\":1234,\"errorString\":\"Sample message 2\",\"recommendations\":[\"recommendation 1\",\"recommendation 2\",\"recommendation 3\"]},{\"errorCode\":1234,\"errorString\":\"Sample message 3\",\"recommendations\":[\"recommendation 1\",\"recommendation 2\",\"recommendation 3\"]}]", detailsInJson)
}

func TestConstants(t *testing.T) {
	suite.Run(t, new(constantsSuite))
}
