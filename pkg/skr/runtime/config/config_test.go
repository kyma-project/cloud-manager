package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"SKR_PROVIDERS": "/env/path",
	})
	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "/env/path", SkrRuntimeConfig.ProvidersDir)
}

func TestConfigFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "skrRuntime.yaml"), []byte(`
providersDir: /some/path/from/file
lockingLeaseDuration: 10s
notificationConcurrency: 16
cyclicConcurrency: 2
cyclicMinInterval: 90s
notificationListenerAddr: ":9090"
gateConflictRetryDelay: 2s
`), 0644)
	assert.NoError(t, err, "error creating key file")

	env := abstractions.NewMockedEnvironment(map[string]string{
		"SKR_PROVIDERS": "/env/path",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "/some/path/from/file", SkrRuntimeConfig.ProvidersDir)
	assert.Equal(t, 10*time.Second, SkrRuntimeConfig.SkrLockingLeaseDuration)
	assert.Equal(t, 16, SkrRuntimeConfig.NotificationConcurrency)
	assert.Equal(t, 2, SkrRuntimeConfig.CyclicConcurrency)
	assert.Equal(t, 90*time.Second, SkrRuntimeConfig.SkrCyclicMinInterval)
	assert.Equal(t, ":9090", SkrRuntimeConfig.NotificationListenerAddr)
	assert.Equal(t, 2*time.Second, SkrRuntimeConfig.SkrGateConflictRetryDelay)
}

func TestConfigDefaults(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 8, SkrRuntimeConfig.NotificationConcurrency)
	assert.Equal(t, 1, SkrRuntimeConfig.CyclicConcurrency)
	assert.Equal(t, 60*time.Second, SkrRuntimeConfig.SkrCyclicMinInterval)
	assert.Equal(t, ":8083", SkrRuntimeConfig.NotificationListenerAddr)
	assert.Equal(t, 1*time.Second, SkrRuntimeConfig.SkrGateConflictRetryDelay)
}

func TestGateConflictRetryDelayFloor(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "skrRuntime.yaml"), []byte(`
gateConflictRetryDelay: 50ms
`), 0644)
	assert.NoError(t, err, "error creating key file")

	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 200*time.Millisecond, SkrRuntimeConfig.SkrGateConflictRetryDelay)
}

func TestNotificationConcurrencyFallsBackToConcurrency(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "skrRuntime.yaml"), []byte(`
concurrency: 5
notificationConcurrency: 0
`), 0644)
	assert.NoError(t, err, "error creating key file")

	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 5, SkrRuntimeConfig.NotificationConcurrency)
}

func TestGateConflictRetryDelayFromEnv(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	env := abstractions.NewMockedEnvironment(map[string]string{
		"SKR_RUNTIME_GATE_CONFLICT_RETRY_DELAY": "3s",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, 3*time.Second, SkrRuntimeConfig.SkrGateConflictRetryDelay)
}
