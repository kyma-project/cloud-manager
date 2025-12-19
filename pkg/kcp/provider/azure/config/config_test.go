package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefault(t *testing.T) {
	AzureConfig = &AzureConfigStruct{}
	cfg := config.NewConfig(abstractions.NewMockedEnvironment(map[string]string{}))
	InitConfig(cfg)
	cfg.Read()

	assert.Empty(t, AzureConfig.DefaultCreds.ClientId)
	assert.Empty(t, AzureConfig.DefaultCreds.ClientSecret)
	assert.Empty(t, AzureConfig.PeeringCreds.ClientId)
	assert.Empty(t, AzureConfig.PeeringCreds.ClientSecret)
	assert.Equal(t, "60s", AzureConfig.FileShareDeletionWait)
	assert.Equal(t, AzureConfig.ClientOptions.Cloud, "AzurePublic")
}

func TestConfigAllFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"AZURE_CLIENT_ID":             "client_id",
		"AZURE_CLIENT_SECRET":         "client_secret",
		"AZURE_PEERING_CLIENT_ID":     "peering_client_id",
		"AZURE_PEERING_CLIENT_SECRET": "peering_client_secret",
		"AZURE_CLIENT_CLOUD":          "AzureChina",
	})
	AzureConfig = &AzureConfigStruct{}
	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "client_id", AzureConfig.DefaultCreds.ClientId)
	assert.Equal(t, "client_secret", AzureConfig.DefaultCreds.ClientSecret)
	assert.Equal(t, "peering_client_id", AzureConfig.PeeringCreds.ClientId)
	assert.Equal(t, "peering_client_secret", AzureConfig.PeeringCreds.ClientSecret)
	assert.Equal(t, "60s", AzureConfig.FileShareDeletionWait)
	assert.Equal(t, AzureConfig.ClientOptions.Cloud, "AzureChina")
}

func TestConfigFromFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "AZURE_CLIENT_ID"), []byte(`file_azure_client_id`), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AZURE_CLIENT_SECRET"), []byte(`file_azure_client_secret`), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AZURE_PEERING_CLIENT_ID"), []byte(`file_azure_peering_client_id`), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AZURE_PEERING_CLIENT_SECRET"), []byte(`file_azure_peering_client_secret`), 0644)
	assert.NoError(t, err, "error creating key file")

	AzureConfig = &AzureConfigStruct{}
	cfg := config.NewConfig(abstractions.NewMockedEnvironment(map[string]string{}))
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "file_azure_client_id", AzureConfig.DefaultCreds.ClientId)
	assert.Equal(t, "file_azure_client_secret", AzureConfig.DefaultCreds.ClientSecret)
	assert.Equal(t, "file_azure_peering_client_id", AzureConfig.PeeringCreds.ClientId)
	assert.Equal(t, "file_azure_peering_client_secret", AzureConfig.PeeringCreds.ClientSecret)
	assert.Equal(t, "60s", AzureConfig.FileShareDeletionWait)
}

func TestConfigFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	err = os.WriteFile(filepath.Join(dir, "azure.yaml"), []byte(`
defaultCreds:
  clientId: file_azure_client_id
  clientSecret: file_azure_client_secret
peeringCreds:
  clientId: file_azure_peering_client_id
  clientSecret: file_azure_peering_client_secret
clientOptions:
  cloud: AzureChina
fileShareDeletionWait: 58s
`), 0644)
	assert.NoError(t, err, "error creating key file")

	AzureConfig = &AzureConfigStruct{}
	cfg := config.NewConfig(abstractions.NewMockedEnvironment(map[string]string{}))
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "file_azure_client_id", AzureConfig.DefaultCreds.ClientId)
	assert.Equal(t, "file_azure_client_secret", AzureConfig.DefaultCreds.ClientSecret)
	assert.Equal(t, "file_azure_peering_client_id", AzureConfig.PeeringCreds.ClientId)
	assert.Equal(t, "file_azure_peering_client_secret", AzureConfig.PeeringCreds.ClientSecret)
	assert.Equal(t, "AzureChina", AzureConfig.ClientOptions.Cloud)
	assert.Equal(t, "58s", AzureConfig.FileShareDeletionWait)
}
