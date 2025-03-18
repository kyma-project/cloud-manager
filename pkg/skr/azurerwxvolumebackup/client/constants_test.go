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

func TestConstants(t *testing.T) {
	suite.Run(t, new(constantsSuite))
}
