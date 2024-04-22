package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type ConfigStruct struct {
	ProvidersDir string `yaml:"providersDir,omitempty" json:"providersDir,omitempty"`
}

var SkrRuntimeConfig = &ConfigStruct{}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"skrRuntime",
		config.Path(
			"providersDir",
			config.DefaultScalar("config/dist/skr/crd/bases/providers"),
			config.SourceEnv("SKR_PROVIDERS"),
		),
		config.SourceFile("skrRuntime.yaml"),
		config.Bind(SkrRuntimeConfig),
	)
}
