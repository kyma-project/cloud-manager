package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AlicloudConfigStruct struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	AccessKeySecret string `json:"accessKeySecret,omitempty" yaml:"accessKeySecret,omitempty"`
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
	)
}
