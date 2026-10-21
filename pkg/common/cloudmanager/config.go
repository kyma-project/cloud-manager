package cloudmanager

import (
	"time"

	"github.com/kyma-project/cloud-manager/pkg/config"
	"k8s.io/client-go/rest"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
)

type ConfigType struct {
	Manager    ManagerConfig `json:"manager" yaml:"manager,omitempty"`
	RestConfig RestConfig    `json:"restConfig" yaml:"restConfig,omitempty"`
}

type ManagerConfig struct {
	Controller ManagerControllerConfig `json:"controller" yaml:"controller,omitempty"`
	Pprof      PprofConfig             `json:"pprof" yaml:"pprof,omitempty"`
}

type ManagerControllerConfig struct {
	MaxConcurrentReconciles int            `json:"maxConcurrentReconciles,omitempty" yaml:"maxConcurrentReconciles,omitempty"`
	GroupKindConcurrency    map[string]int `json:"groupKindConcurrency,omitempty" yaml:"groupKindConcurrency,omitempty"`
	ReconciliationTimeout   time.Duration  `json:"reconciliationTimeout,omitempty" yaml:"reconciliationTimeout,omitempty"`
}

type PprofConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

type RestConfig struct {
	QPS   float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	Burst int     `json:"burst,omitempty" yaml:"burst,omitempty"`
}

func (c *ConfigType) GetManagerControllerConfig() ctrlconfig.Controller {
	result := ctrlconfig.Controller{}
	if c.Manager.Controller.MaxConcurrentReconciles > 0 {
		result.MaxConcurrentReconciles = c.Manager.Controller.MaxConcurrentReconciles
	}
	if c.Manager.Controller.GroupKindConcurrency != nil {
		result.GroupKindConcurrency = c.Manager.Controller.GroupKindConcurrency
	}
	if c.Manager.Controller.ReconciliationTimeout > 0 {
		result.ReconciliationTimeout = c.Manager.Controller.ReconciliationTimeout
	}
	return result
}

func (c *ConfigType) GetManagerPprofBindAddress() string {
	// additional guard beside AfterConfigLoaded(), to ensure empty string is returned when disabled
	if !c.Manager.Pprof.Enabled {
		return ""
	}
	return c.Manager.Pprof.Address
}

func (c *ConfigType) ApplyToRestConfig(rc *rest.Config) {
	if c.RestConfig.QPS > 0 {
		rc.QPS = c.RestConfig.QPS
	}
	if c.RestConfig.Burst > 0 {
		rc.Burst = c.RestConfig.Burst
	}
}

// -------------------------------------------------

var Config = &ConfigType{}

var _ config.AfterConfigLoaded = (*ConfigType)(nil)

func (c *ConfigType) AfterConfigLoaded() {
	if !c.Manager.Pprof.Enabled {
		c.Manager.Pprof.Address = ""
	}
}

func InitConfig(cfg config.Config, instance *ConfigType) {
	if instance == nil {
		instance = Config
	}
	cfg.Path(
		"cloudmanager",
		config.Bind(instance),
		config.SourceFile("cloudmanager.yaml"),

		config.Path(
			"manager.pprof.enabled",
			config.SourceEnv("ENABLE_PPROF"),
		),
		config.Path(
			"manager.pprof.address",
			config.SourceEnv("PPROF_BIND_ADDRESS"),
			config.DefaultScalar(":8082"),
		),

		config.Path(
			"restConfig.qps",
			config.DefaultScalar(float32(50)),
		),
		config.Path(
			"restConfig.burst",
			config.DefaultScalar(100),
		),
	)
}
