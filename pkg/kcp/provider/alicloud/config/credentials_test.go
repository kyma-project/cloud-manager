package config

import (
	"testing"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/assert"
)

// When no assume-role ARN is given, ApplyCredentials must use raw AK/SK and leave
// Credential unset — the fallback path for envtest/mock/single-account setups.
func TestApplyCredentialsRawFallback(t *testing.T) {
	cfg := &openapi.Config{}
	err := ApplyCredentials(cfg, "ak", "sk", "")
	assert.NoError(t, err)
	assert.Equal(t, "ak", tea.StringValue(cfg.AccessKeyId))
	assert.Equal(t, "sk", tea.StringValue(cfg.AccessKeySecret))
	assert.Nil(t, cfg.Credential)
}

// When an assume-role ARN is given, ApplyCredentials must set only Credential and
// leave AccessKeyId/AccessKeySecret unset. The openapi SDK selects raw AK/SK auth
// whenever both are set, so having them non-nil here would silently bypass the
// assume-role credential.
func TestApplyCredentialsAssumeRole(t *testing.T) {
	prev := AlicloudConfig.RoleSessionName
	defer func() { AlicloudConfig.RoleSessionName = prev }()
	AlicloudConfig.RoleSessionName = "cloud-manager"

	cfg := &openapi.Config{}
	err := ApplyCredentials(cfg, "ak", "sk", "acs:ram::123:role/CloudManagerRole")
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Credential)
	assert.Nil(t, cfg.AccessKeyId)
	assert.Nil(t, cfg.AccessKeySecret)
}
