package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type CCEEConfigStruct struct {
	Username      string `json:"username" yaml:"username"`
	Password      string `json:"password" yaml:"password"`
	TlsCaCertPath string `json:"tlsCaCertPath" yaml:"tlsCaCertPath"`
	TlsKeyPath    string `json:"tlsKeyPath" yaml:"tlsKeyPath"`
	TlsCertPath   string `json:"tlsCertPath" yaml:"tlsCertPath"`
}

var CCEEConfig = &CCEEConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"ccee.config",
		config.Bind(CCEEConfig),
		config.Path(
			"username",
			config.SourceEnv("OS_USERNAME"),
			config.SourceFile("OS_USERNAME"),
		),
		config.Path(
			"password",
			config.SourceEnv("OS_PASSWORD"),
			config.SourceFile("OS_PASSWORD"),
		),

		config.Path(
			"tlsCaCertPath",
			config.SourceEnv("OS_CACERT"),
			config.SourceFile("OS_CACERT"),
		),
		config.Path(
			"tlsKeyPath",
			config.SourceEnv("OS_KEY"),
			config.SourceFile("OS_KEY"),
		),
		config.Path(
			"tlsCertPath",
			config.SourceEnv("OS_CERT"),
			config.SourceFile("OS_CERT"),
		),
	)
}
