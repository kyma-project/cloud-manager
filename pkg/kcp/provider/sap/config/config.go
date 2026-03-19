package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type SapConfigStruct struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

	ApplicationCredentialID     string `json:"applicationCredentialID" yaml:"applicationCredentialID"`
	ApplicationCredentialName   string `json:"applicationCredentialName" yaml:"applicationCredentialName"`
	ApplicationCredentialSecret string `json:"applicationCredentialSecret" yaml:"applicationCredentialSecret"`

	FloatingPoolNetwork string `json:"floatingPoolNetwork" yaml:"floatingPoolNetwork"`
	FloatingPoolSubnet  string `json:"floatingPoolSubnet" yaml:"floatingPoolSubnet"`
}

var SapConfig = &SapConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"sap.config",
		config.Bind(SapConfig),
		config.SourceFile("openstack.yaml"),
		config.Path(
			"username",
			config.SourceEnv("OS_USERNAME"),
			config.SourceFile("OS_USERNAME"),
		),
		config.Path(
			"password",
			config.Sensitive(),
			config.SourceEnv("OS_PASSWORD"),
			config.SourceFile("OS_PASSWORD"),
		),
		config.Path(
			"applicationCredentialID",
			config.SourceEnv("OS_APPLICATION_CREDENTIAL_ID"),
			config.SourceFile("OS_APPLICATION_CREDENTIAL_ID"),
		),
		config.Path(
			"applicationCredentialName",
			config.SourceEnv("OS_APPLICATION_CREDENTIAL_NAME"),
			config.SourceFile("OS_APPLICATION_CREDENTIAL_NAME"),
		),
		config.Path(
			"applicationCredentialSecret",
			config.Sensitive(),
			config.SourceEnv("OS_APPLICATION_CREDENTIAL_SECRET"),
			config.SourceFile("OS_APPLICATION_CREDENTIAL_SECRET"),
		),
		config.Path(
			"floatingPoolNetwork",
			config.DefaultScalar("FloatingIP-external-kyma-01"),
			config.SourceEnv("OS_FLOATING_POOL_NETWORK"),
			config.SourceFile("OS_FLOATING_POOL_NETWORK"),
		),
		config.Path(
			"floatingPoolSubnet",
			config.DefaultScalar("FloatingIP-internet-*"),
			config.SourceEnv("OS_FLOATING_POOL_SUBNET"),
			config.SourceFile("OS_FLOATING_POOL_SUBNET"),
		),
	)
}
