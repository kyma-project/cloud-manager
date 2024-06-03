package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AzureConfigStruct struct {
	ClientId     string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
}

var AzureConfig = &AzureConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"azure.config",
		config.Bind(AzureConfig),
		config.Path(
			"clientId",
			config.Sensitive(),
			config.SourceEnv("AZURE_CLIENT_ID"),
			config.SourceFile("AZURE_CLIENT_ID"),
		),
		config.Path(
			"clientSecret",
			config.Sensitive(),
			config.SourceEnv("AZURE_CLIENT_SECRET"),
			config.SourceFile("AZURE_CLIENT_SECRET"),
		),
	)
}
