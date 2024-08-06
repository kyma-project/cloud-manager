package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AzureConfigStruct struct {
	ClientId               string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret           string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	VpcPeeringClientId     string `json:"vpcPeeringClientId,omitempty" yaml:"vpcPeeringClientId,omitempty"`
	VpcPeeringClientSecret string `json:"vpcPeeringClientSecret,omitempty" yaml:"vpcPeeringClientSecret,omitempty"`
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
		config.Path(
			"vpcPeeringClientId",
			config.Sensitive(),
			config.SourceEnv("AZURE_VPC_PEERING_CLIENT_ID"),
			config.SourceFile("AZURE_VPC_PEERING_CLIENT_ID"),
		),
		config.Path(
			"vpcPeeringClientSecret",
			config.Sensitive(),
			config.SourceEnv("AZURE_VPC_PEERING_CLIENT_SECRET"),
			config.SourceFile("AZURE_VPC_PEERING_CLIENT_SECRET"),
		),
	)
}
