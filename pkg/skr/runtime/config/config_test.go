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
}
