package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigAllFromEnv(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"AWS_ACCESS_KEY_ID":                 "key",
		"AWS_SECRET_ACCESS_KEY":             "secret",
		"AWS_ROLE_NAME":                     "role",
		"AWS_ELASTICACHE_ACCESS_KEY_ID":     "elasti-key",
		"AWS_ELASTICACHE_SECRET_ACCESS_KEY": "elasti-secret",
		"AWS_ELASTICACHE_ROLE_NAME":         "elasti-role",
	})
	cfg := config.NewConfig(env)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "key", AwsConfig.Default.AccessKeyId)
	assert.Equal(t, "secret", AwsConfig.Default.SecretAccessKey)
	assert.Equal(t, "role", AwsConfig.Default.AssumeRoleName)
	assert.Equal(t, "elasti-key", AwsConfig.ElastiCache.AccessKeyId)
	assert.Equal(t, "elasti-secret", AwsConfig.ElastiCache.SecretAccessKey)
	assert.Equal(t, "elasti-role", AwsConfig.ElastiCache.AssumeRoleName)
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

	assert.Equal(t, "key222", AwsConfig.Default.AccessKeyId)
	assert.Equal(t, "secret222", AwsConfig.Default.SecretAccessKey)
	assert.Equal(t, "role", AwsConfig.Default.AssumeRoleName)
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

	err = os.WriteFile(filepath.Join(dir, "AWS_ELASTICACHE_ACCESS_KEY_ID"), []byte("key333"), 0644)
	assert.NoError(t, err, "error creating key file")
	err = os.WriteFile(filepath.Join(dir, "AWS_ELASTICACHE_SECRET_ACCESS_KEY"), []byte("secret333"), 0644)
	assert.NoError(t, err, "error creating secret file")
	err = os.WriteFile(filepath.Join(dir, "AWS_ELASTICACHE_ROLE_NAME"), []byte("role333"), 0644)
	assert.NoError(t, err, "error creating fole file")

	env := abstractions.NewMockedEnvironment(map[string]string{
		"AWS_ACCESS_KEY_ID":                 "key",
		"AWS_SECRET_ACCESS_KEY":             "secret",
		"AWS_ROLE_NAME":                     "role",
		"AWS_ELASTICACHE_ACCESS_KEY_ID":     "elasti-key",
		"AWS_ELASTICACHE_SECRET_ACCESS_KEY": "elasti-secret",
		"AWS_ELASTICACHE_ROLE_NAME":         "elasti-role",
	})
	cfg := config.NewConfig(env)
	cfg.BaseDir(dir)
	InitConfig(cfg)
	cfg.Read()

	assert.Equal(t, "key222", AwsConfig.Default.AccessKeyId)
	assert.Equal(t, "secret222", AwsConfig.Default.SecretAccessKey)
	assert.Equal(t, "role222", AwsConfig.Default.AssumeRoleName)
	assert.Equal(t, "key333", AwsConfig.ElastiCache.AccessKeyId)
	assert.Equal(t, "secret333", AwsConfig.ElastiCache.SecretAccessKey)
	assert.Equal(t, "role333", AwsConfig.ElastiCache.AssumeRoleName)
}
