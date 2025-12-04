package config

import (
	config2 "github.com/kyma-project/cloud-manager/pkg/config"
	"time"
)

type GcpConfigStruct struct {
	GcpRetryWaitTime         time.Duration `mapstructure:"retryWaitTime,omitempty"`
	GcpOperationWaitTime     time.Duration `mapstructure:"operationWaitTime,omitempty"`
	GcpApiTimeout            time.Duration `mapstructure:"apiTimeout,omitempty"`
	GcpCapacityCheckInterval time.Duration `mapstructure:"capacityCheckInterval,omitempty"`
	CredentialsFile          string        `mapstructure:"credentialsFile,omitempty"`
	PeeringCredentialsFile   string        `mapstructure:"peeringCredentialsFile,omitempty"`
}

func InitConfig(cfg config2.Config) {
	cfg.Path(
		"gcpConfig",
		config2.Bind(GcpConfig),
		config2.SourceFile("gcpConfig.yaml"),
		config2.Path(
			"retryWaitTime",
			config2.DefaultScalar("5s"),
			config2.SourceEnv("GCP_RETRY_WAIT_DURATION"),
		),
		config2.Path(
			"operationWaitTime",
			config2.DefaultScalar("5s"),
			config2.SourceEnv("GCP_OPERATION_WAIT_DURATION"),
		),
		config2.Path(
			"apiTimeout",
			config2.DefaultScalar("8s"),
			config2.SourceEnv("GCP_API_TIMEOUT_DURATION"),
		),
		config2.Path(
			"capacityCheckInterval",
			config2.DefaultScalar("1h"),
			config2.SourceEnv("GCP_CAPACITY_CHECK_INTERVAL"),
		),
		config2.Path("credentialsFile",
			config2.SourceEnv("GCP_SA_JSON_KEY_PATH"),
			config2.SourceFile("GCP_SA_JSON_KEY_PATH"),
		),
		config2.Path("peeringCredentialsFile",
			config2.SourceEnv("GCP_VPC_PEERING_KEY_PATH"),
			config2.SourceFile("GCP_VPC_PEERING_KEY_PATH"),
		),
	)
}

var GcpConfig = &GcpConfigStruct{}
