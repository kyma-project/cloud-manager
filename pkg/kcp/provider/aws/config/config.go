package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AwsConfigStruct struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	AssumeRoleName  string `json:"assumeRoleName,omitempty" yaml:"assumeRoleName,omitempty"`
}

var AwsConfig = &AwsConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"aws.config",
		config.Bind(AwsConfig),
		config.Path(
			"accessKeyId",
			config.SourceEnv("AWS_ACCESS_KEY_ID"),
			config.SourceFile("AWS_ACCESS_KEY_ID"),
		),
		config.Path(
			"secretAccessKey",
			config.Sensitive(),
			config.SourceEnv("AWS_SECRET_ACCESS_KEY"),
			config.SourceFile("AWS_SECRET_ACCESS_KEY"),
		),
		config.Path(
			"assumeRoleName",
			config.SourceEnv("AWS_ROLE_NAME"),
			config.SourceFile("AWS_ROLE_NAME"),
		),
	)

}
