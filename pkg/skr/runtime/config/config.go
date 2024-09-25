package config

import (
	"github.com/kyma-project/cloud-manager/pkg/config"
	"time"
)

type ConfigStruct struct {
	SkrLockingLeaseDuration time.Duration

	ProvidersDir         string `yaml:"providersDir,omitempty" json:"providersDir,omitempty"`
	Concurrency          int    `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	LockingLeaseDuration string `yaml:"lockingLeaseDuration,omitempty" json:"lockingLeaseDuration,omitempty"`
}

func (c *ConfigStruct) AfterConfigLoaded() {
	if c.Concurrency <= 0 {
		c.Concurrency = 1
	}
	if c.Concurrency > 100 {
		c.Concurrency = 100
	}
	c.SkrLockingLeaseDuration = GetDuration(c.LockingLeaseDuration, 10*time.Minute)

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
		config.Path(
			"lockingLeaseDuration",
			config.DefaultScalar("600s"),
		),
		config.SourceFile("skrRuntime.yaml"),
		config.Bind(SkrRuntimeConfig),
	)
}

func GetDuration(value string, defaultValue time.Duration) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
