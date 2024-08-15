package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AzureCreds struct {
	ClientId     string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
}
type AzureConfigStruct struct {
	DefaultCreds AzureCreds `json:"defaultCreds" yaml:"defaultCreds"`
	PeeringCreds AzureCreds `json:"peeringCreds" yaml:"peeringCreds"`
}

var AzureConfig = &AzureConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"azure.config",
		config.Bind(AzureConfig),
		config.Path(
			"defaultCreds.clientId",
			config.Sensitive(),
			config.SourceEnv("AZURE_CLIENT_ID"),
			config.SourceFile("AZURE_CLIENT_ID"),
		),
		config.Path(
			"defaultCreds.clientSecret",
			config.Sensitive(),
			config.SourceEnv("AZURE_CLIENT_SECRET"),
			config.SourceFile("AZURE_CLIENT_SECRET"),
		),
		config.Path(
			"peeringCreds.clientId",
			config.Sensitive(),
			config.SourceEnv("AZURE_PEERING_CLIENT_ID"),
			config.SourceFile("AZURE_PEERING_CLIENT_ID"),
			config.SourceFile("peering/AZURE_CLIENT_ID"),
		),
		config.Path(
			"peeringCreds.clientSecret",
			config.Sensitive(),
			config.SourceEnv("AZURE_PEERING_CLIENT_SECRET"),
			config.SourceFile("AZURE_PEERING_CLIENT_SECRET"),
			config.SourceFile("peering/AZURE_CLIENT_SECRET"),
		),
	)
}
