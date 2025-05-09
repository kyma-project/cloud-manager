package config

import (
	"time"

	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AzureCreds struct {
	ClientId     string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
}
type AzureConfigStruct struct {
	DefaultCreds          AzureCreds `json:"defaultCreds" yaml:"defaultCreds"`
	PeeringCreds          AzureCreds `json:"peeringCreds" yaml:"peeringCreds"`
	FileShareDeletionWait string     `json:"fileShareDeletionWait" yaml:"fileShareDeletionWait"`

	AzureFileShareDeletionWaitDuration time.Duration
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
		config.Path(
			"fileShareDeletionWait",
			config.DefaultScalar("60s"),
			config.SourceEnv("AZURE_FILE_SHARE_DELETION_WAIT"),
		),
	)
}

func (c *AzureConfigStruct) AfterConfigLoaded() {
	c.AzureFileShareDeletionWaitDuration = GetDuration(c.FileShareDeletionWait, time.Second*60)
}

func GetDuration(value string, defaultValue time.Duration) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
