package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AwsConfigStruct struct {
	Default AwsCreds `json:"default" yaml:"default"`
	Peering AwsCreds `json:"peering" yaml:"peering"`
}

var AwsConfig = &AwsConfigStruct{}

type AwsCreds struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	AssumeRoleName  string `json:"assumeRoleName,omitempty" yaml:"assumeRoleName,omitempty"`
}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"aws.config",
		config.Bind(AwsConfig),
		config.Path(
			"default.accessKeyId",
			config.SourceEnv("AWS_ACCESS_KEY_ID"),
			config.SourceFile("AWS_ACCESS_KEY_ID"),
		),
		config.Path(
			"default.secretAccessKey",
			config.Sensitive(),
			config.SourceEnv("AWS_SECRET_ACCESS_KEY"),
			config.SourceFile("AWS_SECRET_ACCESS_KEY"),
		),
		config.Path(
			"default.assumeRoleName",
			config.SourceEnv("AWS_ROLE_NAME"),
			config.SourceFile("AWS_ROLE_NAME"),
		),
		config.Path(
			"peering.accessKeyId",
			config.SourceEnv("AWS_PEERING_ACCESS_KEY_ID"),
			config.SourceFile("AWS_PEERING_ACCESS_KEY_ID"),
		),
		config.Path(
			"peering.secretAccessKey",
			config.Sensitive(),
			config.SourceEnv("AWS_PEERING_SECRET_ACCESS_KEY"),
			config.SourceFile("AWS_PEERING_SECRET_ACCESS_KEY"),
		),
		config.Path(
			"peering.assumeRoleName",
			config.SourceEnv("AWS_PEERING_ROLE_NAME"),
			config.SourceFile("AWS_PEERING_ROLE_NAME"),
		),
	)

}
