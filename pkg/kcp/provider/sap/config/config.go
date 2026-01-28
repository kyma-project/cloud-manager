package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type SapConfigStruct struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

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
			"floatingPoolNetwork",
			config.DefaultScalar("FloatingIP-external-kyma-01"),
		),
		config.Path(
			"floatingPoolSubnet",
			config.DefaultScalar("FloatingIP-internet-*"),
		),
	)
}
