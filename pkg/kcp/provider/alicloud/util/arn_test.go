package util

import (
	"testing"

	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
	"github.com/stretchr/testify/assert"
)

func TestRoleArn(t *testing.T) {
	assert.Equal(t, "acs:ram::123456789:role/CloudManagerRole", RoleArn("123456789", "CloudManagerRole"))
}

func TestRoleArnDefault(t *testing.T) {
	prev := alicloudconfig.AlicloudConfig.AssumeRoleName
	defer func() { alicloudconfig.AlicloudConfig.AssumeRoleName = prev }()
	alicloudconfig.AlicloudConfig.AssumeRoleName = "CloudManagerRole"

	assert.Equal(t, "acs:ram::123456789:role/CloudManagerRole", RoleArnDefault("123456789"))
	// Empty account id yields empty ARN so callers fall back to raw AK/SK.
	assert.Equal(t, "", RoleArnDefault(""))
}
