package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestCceeConfig(t *testing.T) {

	const (
		un   = "un"
		pw   = "pw"
		ca   = "ca"
		key  = "key"
		cert = "cert"
	)

	values := map[string]string{
		"OS_USERNAME": un,
		"OS_PASSWORD": pw,
		"OS_CACERT":   ca,
		"OS_KEY":      key,
		"OS_CERT":     cert,
	}

	t.Run("All from env", func(t *testing.T) {
		env := abstractions.NewMockedEnvironment(values)
		cfg := config.NewConfig(env)
		InitConfig(cfg)
		cfg.Read()

		assert.Equal(t, un, CCEEConfig.Username)
		assert.Equal(t, pw, CCEEConfig.Password)
		assert.Equal(t, ca, CCEEConfig.TlsCaCertPath)
		assert.Equal(t, key, CCEEConfig.TlsKeyPath)
		assert.Equal(t, cert, CCEEConfig.TlsCertPath)
	})

	t.Run("All from file", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "cloud-manager-ccee-config")
		assert.NoError(t, err, "error creating tmp dir")
		defer func() {
			_ = os.RemoveAll(dir)
		}()
		for k, v := range values {
			err = os.WriteFile(filepath.Join(dir, k), []byte(v), 0644)
			assert.NoError(t, err, "error creating file %s", k)
		}

		fakeValues := map[string]string{}
		for k, v := range values {
			fakeValues[k] = v + v + v // not the same as in the file
		}
		env := abstractions.NewMockedEnvironment(fakeValues)

		cfg := config.NewConfig(env)
		cfg.BaseDir(dir)
		InitConfig(cfg)
		cfg.Read()

		assert.Equal(t, un, CCEEConfig.Username)
		assert.Equal(t, pw, CCEEConfig.Password)
		assert.Equal(t, ca, CCEEConfig.TlsCaCertPath)
		assert.Equal(t, key, CCEEConfig.TlsKeyPath)
		assert.Equal(t, cert, CCEEConfig.TlsCertPath)
	})

}
