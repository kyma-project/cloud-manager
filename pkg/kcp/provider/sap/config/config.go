package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type SapConfigStruct struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

var SapConfig = &SapConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"sap.config",
		config.Bind(SapConfig),
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
	)
}
