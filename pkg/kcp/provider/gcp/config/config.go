package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
	"time"
)

type GcpConfigStruct struct {
	GcpRetryWaitTime         time.Duration `mapstructure:"retryWaitTime,omitempty"`
	GcpOperationWaitTime     time.Duration `mapstructure:"operationWaitTime,omitempty"`
	GcpApiTimeout            time.Duration `mapstructure:"apiTimeout,omitempty"`
	GcpCapacityCheckInterval time.Duration `mapstructure:"capacityCheckInterval,omitempty"`
	ClientRenewDuration      time.Duration `mapstructure:"clientRenewDuration,omitempty"`
	CredentialsFile          string        `mapstructure:"credentialsFile,omitempty"`
	PeeringCredentialsFile   string        `mapstructure:"peeringCredentialsFile,omitempty"`
}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"gcpConfig",
		config.Bind(GcpConfig),
		config.SourceFile("gcpConfig.yaml"),
		config.Path(
			"retryWaitTime",
			config.DefaultScalar("5s"),
			config.SourceEnv("GCP_RETRY_WAIT_DURATION"),
		),
		config.Path(
			"operationWaitTime",
			config.DefaultScalar("5s"),
			config.SourceEnv("GCP_OPERATION_WAIT_DURATION"),
		),
		config.Path(
			"apiTimeout",
			config.DefaultScalar("8s"),
			config.SourceEnv("GCP_API_TIMEOUT_DURATION"),
		),
		config.Path(
			"capacityCheckInterval",
			config.DefaultScalar("1h"),
			config.SourceEnv("GCP_CAPACITY_CHECK_INTERVAL"),
		),
		config.Path(
			"clientRenewDuration",
			config.DefaultScalar("5m"),
			config.SourceEnv("GCP_CLIENT_RENEW_DURATION"),
		),
		config.Path("credentialsFile",
			config.SourceEnv("GCP_SA_JSON_KEY_PATH"),
			config.SourceFile("GCP_SA_JSON_KEY_PATH"),
		),
		config.Path("peeringCredentialsFile",
			config.SourceEnv("GCP_VPC_PEERING_KEY_PATH"),
			config.SourceFile("GCP_VPC_PEERING_KEY_PATH"),
		),
	)
}

var GcpConfig = &GcpConfigStruct{}
