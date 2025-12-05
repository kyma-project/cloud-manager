package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestFromEnv(t *testing.T) {

	GcpConfig = &GcpConfigStruct{}

	env := abstractions.NewMockedEnvironment(map[string]string{
		"GCP_RETRY_WAIT_DURATION":     "123s",
		"GCP_OPERATION_WAIT_DURATION": "124s",
		"GCP_API_TIMEOUT_DURATION":    "125s",
		"GCP_CAPACITY_CHECK_INTERVAL": "126s",
		"GCP_SA_JSON_KEY_PATH":        "/tmp/credentials",
		"GCP_VPC_PEERING_KEY_PATH":    "/tmp/peering/credentials",
	})

	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 123*time.Second, GcpConfig.GcpRetryWaitTime)

	assert.Equal(t, 124*time.Second, GcpConfig.GcpOperationWaitTime)

	assert.Equal(t, 125*time.Second, GcpConfig.GcpApiTimeout)

	assert.Equal(t, 126*time.Second, GcpConfig.GcpCapacityCheckInterval)

	assert.Equal(t, "/tmp/credentials", GcpConfig.CredentialsFile)

	assert.Equal(t, "/tmp/peering/credentials", GcpConfig.PeeringCredentialsFile)
}

func TestEmpty(t *testing.T) {
	// must instantiate a new one to isolate from previous tests
	GcpConfig = &GcpConfigStruct{}

	env := abstractions.NewMockedEnvironment(map[string]string{})

	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 5*time.Second, GcpConfig.GcpRetryWaitTime)

	assert.Equal(t, 5*time.Second, GcpConfig.GcpOperationWaitTime)

	assert.Equal(t, 8*time.Second, GcpConfig.GcpApiTimeout)

	assert.Equal(t, time.Hour, GcpConfig.GcpCapacityCheckInterval)

	assert.Equal(t, 5*time.Minute, GcpConfig.ClientRenewDuration)

	assert.Empty(t, GcpConfig.CredentialsFile)

	assert.Empty(t, GcpConfig.PeeringCredentialsFile)
}

func TestFromFile(t *testing.T) {

	t.Context()
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "gcpConfig.yaml"), []byte(`
retryWaitTime: 1s
operationWaitTime: 2s
apiTimeout: 3s
capacityCheckInterval: 4h
clientRenewDuration: 10m
credentialsFile: /tmp/credentials
peeringCredentialsFile: /tmp/peering/credentials
`), 0644)
	assert.NoError(t, err, "error creating key file")

	GcpConfig = &GcpConfigStruct{}

	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, time.Second, GcpConfig.GcpRetryWaitTime)
	assert.Equal(t, 2*time.Second, GcpConfig.GcpOperationWaitTime)
	assert.Equal(t, 3*time.Second, GcpConfig.GcpApiTimeout)
	assert.Equal(t, 4*time.Hour, GcpConfig.GcpCapacityCheckInterval)
	assert.Equal(t, 10*time.Minute, GcpConfig.ClientRenewDuration)
	assert.Equal(t, "/tmp/credentials", GcpConfig.CredentialsFile)
	assert.Equal(t, "/tmp/peering/credentials", GcpConfig.PeeringCredentialsFile)
}

func TestEnvOverrideFile(t *testing.T) {

	t.Context()
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "gcpConfig.yaml"), []byte(`
retryWaitTime: 1s
operationWaitTime: 2s
apiTimeout: 3s
capacityCheckInterval: 4h
clientRenewDuration: 7m
credentialsFile: /tmp/credentials
peeringCredentialsFile: /tmp/peering/credentials
`), 0644)
	assert.NoError(t, err, "error creating key file")

	GcpConfig = &GcpConfigStruct{}

	env := abstractions.NewMockedEnvironment(map[string]string{
		"GCP_SA_JSON_KEY_PATH":        "/env/credentials",
		"GCP_RETRY_WAIT_DURATION":     "10s",
		"GCP_OPERATION_WAIT_DURATION": "20s",
		"GCP_API_TIMEOUT_DURATION":    "30s",
		"GCP_CAPACITY_CHECK_INTERVAL": "40h",
		"GCP_CLIENT_RENEW_DURATION":   "20m",
	})

	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 10*time.Second, GcpConfig.GcpRetryWaitTime)
	assert.Equal(t, 20*time.Second, GcpConfig.GcpOperationWaitTime)
	assert.Equal(t, 30*time.Second, GcpConfig.GcpApiTimeout)
	assert.Equal(t, 40*time.Hour, GcpConfig.GcpCapacityCheckInterval)
	assert.Equal(t, 20*time.Minute, GcpConfig.ClientRenewDuration)
	assert.Equal(t, "/env/credentials", GcpConfig.CredentialsFile)
	assert.Equal(t, "/tmp/peering/credentials", GcpConfig.PeeringCredentialsFile)
}
