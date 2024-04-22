package config

import (
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigAllFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"AWS_ACCESS_KEY_ID":     "key",
		"AWS_SECRET_ACCESS_KEY": "secret",
		"AWS_ROLE_NAME":         "role",
	})
	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "key", AwsConfig.AccessKeyId)
	assert.Equal(t, "secret", AwsConfig.SecretAccessKey)
	assert.Equal(t, "role", AwsConfig.AssumeRoleName)
}

func TestConfigCredentialsFromFileRoleFromEnv(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "AWS_ACCESS_KEY_ID"), []byte("key222"), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AWS_SECRET_ACCESS_KEY"), []byte("secret222"), 0644)
	assert.NoError(t, err, "error creating secret file")

	env := abstractions.NewMockedEnvironment(map[string]string{
		"AWS_ACCESS_KEY_ID":     "key",
		"AWS_SECRET_ACCESS_KEY": "secret",
		"AWS_ROLE_NAME":         "role",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "key222", AwsConfig.AccessKeyId)
	assert.Equal(t, "secret222", AwsConfig.SecretAccessKey)
	assert.Equal(t, "role", AwsConfig.AssumeRoleName)
}

func TestAllFromFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloud-manager-config")
	assert.NoError(t, err, "error creating tmp dir")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	err = os.WriteFile(filepath.Join(dir, "AWS_ACCESS_KEY_ID"), []byte("key222"), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AWS_SECRET_ACCESS_KEY"), []byte("secret222"), 0644)
	assert.NoError(t, err, "error creating secret file")
	err = os.WriteFile(filepath.Join(dir, "AWS_ROLE_NAME"), []byte("role222"), 0644)
	assert.NoError(t, err, "error creating fole file")

	env := abstractions.NewMockedEnvironment(map[string]string{
		"AWS_ACCESS_KEY_ID":     "key",
		"AWS_SECRET_ACCESS_KEY": "secret",
		"AWS_ROLE_NAME":         "role",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "key222", AwsConfig.AccessKeyId)
	assert.Equal(t, "secret222", AwsConfig.SecretAccessKey)
	assert.Equal(t, "role222", AwsConfig.AssumeRoleName)
}
