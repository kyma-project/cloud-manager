package config

import (
	"time"

	"github.com/kyma-project/cloud-manager/pkg/config"
)

type ConfigStruct struct {
	SkrLockingLeaseDuration   time.Duration
	SkrCyclicMinInterval      time.Duration
	SkrGateConflictRetryDelay time.Duration

	ProvidersDir         string `yaml:"providersDir,omitempty" json:"providersDir,omitempty"`
	Concurrency          int    `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	LockingLeaseDuration string `yaml:"lockingLeaseDuration,omitempty" json:"lockingLeaseDuration,omitempty"`

	// Two-sleeve looper configuration.
	// NotificationConcurrency sizes the (large) notification worker pool driven by
	// runtime-watcher events; CyclicConcurrency sizes the (small) background round-robin pool.
	NotificationConcurrency  int    `yaml:"notificationConcurrency,omitempty" json:"notificationConcurrency,omitempty"`
	CyclicConcurrency        int    `yaml:"cyclicConcurrency,omitempty" json:"cyclicConcurrency,omitempty"`
	CyclicMinInterval        string `yaml:"cyclicMinInterval,omitempty" json:"cyclicMinInterval,omitempty"`
	NotificationListenerAddr string `yaml:"notificationListenerAddr,omitempty" json:"notificationListenerAddr,omitempty"`
	GateConflictRetryDelay   string `yaml:"gateConflictRetryDelay,omitempty" json:"gateConflictRetryDelay,omitempty"`
}

func (c *ConfigStruct) AfterConfigLoaded() {
	if c.Concurrency <= 0 {
		c.Concurrency = 1
	}
	if c.Concurrency > 100 {
		c.Concurrency = 100
	}
	c.SkrLockingLeaseDuration = GetDuration(c.LockingLeaseDuration, 10*time.Minute)

	// Backwards-compat: deployments that only set the legacy Concurrency (and not the new
	// NotificationConcurrency) get their value carried over to the notification pool so the
	// primary user-facing sleeve keeps a sane size.
	if c.NotificationConcurrency <= 0 {
		c.NotificationConcurrency = c.Concurrency
	}
	if c.NotificationConcurrency > 100 {
		c.NotificationConcurrency = 100
	}
	if c.CyclicConcurrency <= 0 {
		c.CyclicConcurrency = 1
	}
	if c.CyclicConcurrency > 100 {
		c.CyclicConcurrency = 100
	}

	c.SkrCyclicMinInterval = GetDuration(c.CyclicMinInterval, 60*time.Second)

	// Floor-clamp the gate conflict retry delay to avoid a busy-requeue loop on misconfiguration.
	c.SkrGateConflictRetryDelay = GetDuration(c.GateConflictRetryDelay, 1*time.Second)
	if c.SkrGateConflictRetryDelay < 200*time.Millisecond {
		c.SkrGateConflictRetryDelay = 200 * time.Millisecond
	}

	if c.NotificationListenerAddr == "" {
		c.NotificationListenerAddr = ":8083"
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
			config.SourceEnv("SKR_RUNTIME_CONCURRENCY"),
		),
		config.Path(
			"lockingLeaseDuration",
			config.DefaultScalar("600s"),
		),
		config.Path(
			"notificationConcurrency",
			config.DefaultScalar(8),
			config.SourceEnv("SKR_RUNTIME_NOTIFICATION_CONCURRENCY"),
		),
		config.Path(
			"cyclicConcurrency",
			config.DefaultScalar(1),
			config.SourceEnv("SKR_RUNTIME_CYCLIC_CONCURRENCY"),
		),
		config.Path(
			"cyclicMinInterval",
			config.DefaultScalar("60s"),
		),
		config.Path(
			"notificationListenerAddr",
			config.DefaultScalar(":8083"),
			config.SourceEnv("SKR_RUNTIME_NOTIFICATION_LISTENER_ADDR"),
		),
		config.Path(
			"gateConflictRetryDelay",
			config.DefaultScalar("1s"),
			config.SourceEnv("SKR_RUNTIME_GATE_CONFLICT_RETRY_DELAY"),
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
