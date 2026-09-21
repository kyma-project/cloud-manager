package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AlicloudConfigStruct struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	AccessKeySecret string `json:"accessKeySecret,omitempty" yaml:"accessKeySecret,omitempty"`
	// AssumeRoleName is the name of the per-account RAM role Cloud Manager assumes
	// in each runtime AliCloud account (ARN acs:ram::<accountId>:role/<AssumeRoleName>).
	AssumeRoleName string `json:"assumeRoleName,omitempty" yaml:"assumeRoleName,omitempty"`
	// RoleSessionName identifies the assume-role session in AliCloud ActionTrail.
	RoleSessionName string `json:"roleSessionName,omitempty" yaml:"roleSessionName,omitempty"`
}

var AlicloudConfig = &AlicloudConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"alicloud.config",
		config.Bind(AlicloudConfig),
		config.SourceFile("alicloud.yaml"),
		config.Path(
			"accessKeyId",
			config.SourceEnv("ALICLOUD_ACCESS_KEY"),
			config.SourceFile("ALICLOUD_ACCESS_KEY"),
		),
		config.Path(
			"accessKeySecret",
			config.Sensitive(),
			config.SourceEnv("ALICLOUD_SECRET_KEY"),
			config.SourceFile("ALICLOUD_SECRET_KEY"),
		),
		config.Path(
			"assumeRoleName",
			config.DefaultScalar("CloudManagerRole"),
			config.SourceEnv("ALICLOUD_ROLE_NAME"),
		),
		config.Path(
			"roleSessionName",
			config.DefaultScalar("cloud-manager"),
			config.SourceEnv("ALICLOUD_ROLE_SESSION_NAME"),
		),
	)
}
