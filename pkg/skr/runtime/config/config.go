package config

import "github.com/kyma-project/cloud-manager/pkg/config"

type ConfigStruct struct {
	ProvidersDir string `yaml:"providersDir,omitempty" json:"providersDir,omitempty"`
	Concurrency  int    `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
}

func (c *ConfigStruct) AfterConfigLoaded() {
	if c.Concurrency <= 0 {
		c.Concurrency = 1
	}
	if c.Concurrency > 100 {
		c.Concurrency = 100
	}
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
		config.Path(
			"concurrency",
			config.DefaultScalar(1),
		),
		config.SourceFile("skrRuntime.yaml"),
		config.Bind(SkrRuntimeConfig),
	)
}
