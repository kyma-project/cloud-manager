package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"GARDENER_NAMESPACE": "ns-env",
	})
	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "ns-env", ScopeConfig.GardenerNamespace)
}

func TestConfigFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "scope.yaml"), []byte(`
gardenerNamespace: ns-file
`), 0644)
	assert.NoError(t, err, "error creating key file")

	env := abstractions.NewMockedEnvironment(map[string]string{
		"GARDENER_NAMESPACE": "ns-env",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "ns-file", ScopeConfig.GardenerNamespace)
}
