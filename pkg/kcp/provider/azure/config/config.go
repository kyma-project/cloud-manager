package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AzureConfigStruct struct {
	ClientId       string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret   string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	SubscriptionId string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantId       string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
}

var AzureConfig = &AzureConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"azure.config",
		config.Bind(AzureConfig),
		config.Path(
			"clientId",
			config.Sensitive(),
			config.SourceEnv("AZ_CLIENT_ID"),
			config.SourceFile("AZ_CLIENT_ID"),
		),
		config.Path(
			"clientSecret",
			config.Sensitive(),
			config.SourceEnv("AZ_CLIENT_SECRET"),
			config.SourceFile("AZ_CLIENT_SECRET"),
		),
		config.Path(
			"subscriptionId",
			config.Sensitive(),
			config.SourceEnv("AZ_SUBSCRIPTION_ID"),
			config.SourceFile("AZ_SUBSCRIPTION_ID"),
		),
		config.Path(
			"tenantId",
			config.Sensitive(),
			config.SourceEnv("AZ_TENANT_ID"),
			config.SourceFile("AZ_TENANT_ID"),
		),
	)
}
