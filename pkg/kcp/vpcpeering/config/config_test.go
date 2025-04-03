package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	pkgconfig "github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"PEERING_NETWORK_TAG": "e2e",
	})
	cfg := pkgconfig.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "e2e", VpcPeeringConfig.NetworkTag)
	assert.Equal(t, true, VpcPeeringConfig.RouteAsociatedCidrBlocks)
}

func TestConfigFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "vpcpeering.yaml"), []byte(`
networkTag: e2e
`), 0644)
	assert.NoError(t, err, "error creating key file")

	env := abstractions.NewMockedEnvironment(map[string]string{})
	cfg := pkgconfig.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, true, VpcPeeringConfig.RouteAsociatedCidrBlocks)
}
