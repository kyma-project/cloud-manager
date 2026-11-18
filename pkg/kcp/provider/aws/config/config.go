package config

import (
	"strings"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/config"
)

// EfsOrphanRecoveryConfig configures the disposable EFS orphan-recovery garbage collector.
// It is bound from the aws.yaml config file only (no env), so the incident-specific shootId,
// awsAccount and region never appear as code literals. Disposable: removed after close-out.
type EfsOrphanRecoveryConfig struct {
	Mode        string        `json:"mode" yaml:"mode"` // off | dry-run | on
	ShootId     string        `json:"shootId" yaml:"shootId"`
	AwsAccount  string        `json:"awsAccount" yaml:"awsAccount"`
	Region      string        `json:"region" yaml:"region"`
	DeleteSleep time.Duration `json:"deleteSleep" yaml:"deleteSleep"`
}

const (
	EfsOrphanRecoveryModeOff    = "off"
	EfsOrphanRecoveryModeDryRun = "dry-run"
	EfsOrphanRecoveryModeOn     = "on"
)

type AwsConfigStruct struct {
	ArnPartition                  string                                               `json:"arnPartition" yaml:"arnPartition"`
	Default                       AwsCreds                                             `json:"default" yaml:"default"`
	Peering                       AwsCreds                                             `json:"peering" yaml:"peering"`
	BackupRoleName                string                                               `json:"backupRoleName" yaml:"backupRoleName"`
	EfsCapacityCheckInterval      time.Duration                                        `json:"efsCapacityCheckInterval" yaml:"efsCapacityCheckInterval"`
	EfsOrphanRecovery             EfsOrphanRecoveryConfig                              `json:"efsOrphanRecovery" yaml:"efsOrphanRecovery"`
	RedisInstanceTierMachineTypes map[cloudresourcesv1beta1.AwsRedisTier]string        `json:"redisInstanceTierMachineTypes" yaml:"redisInstanceTierMachineTypes"`
	RedisClusterTierMachineTypes  map[cloudresourcesv1beta1.AwsRedisClusterTier]string `json:"redisClusterTierMachineTypes" yaml:"redisClusterTierMachineTypes"`
}

var AwsConfig = &AwsConfigStruct{}

type AwsCreds struct {
	AccessKeyId     string `json:"accessKeyId,omitempty" yaml:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	AssumeRoleName  string `json:"assumeRoleName,omitempty" yaml:"assumeRoleName,omitempty"`
}

func InitConfig(cfg config.Config) {
	cfg.Path(
		"aws.config",
		config.Bind(AwsConfig),
		config.SourceFile("aws.yaml"),
		config.Path(
			"arnPartition",
			config.SourceEnv("AWS_PARTITION"),
			config.SourceFile("AWS_PARTITION"),
			config.DefaultScalar("aws"),
		),
		config.Path(
			"default.accessKeyId",
			config.SourceEnv("AWS_ACCESS_KEY_ID"),
			config.SourceFile("AWS_ACCESS_KEY_ID"),
		),
		config.Path(
			"default.secretAccessKey",
			config.Sensitive(),
			config.SourceEnv("AWS_SECRET_ACCESS_KEY"),
			config.SourceFile("AWS_SECRET_ACCESS_KEY"),
		),
		config.Path(
			"default.assumeRoleName",
			config.SourceEnv("AWS_ROLE_NAME"),
			config.SourceFile("AWS_ROLE_NAME"),
		),
		config.Path(
			"peering.accessKeyId",
			config.SourceEnv("AWS_PEERING_ACCESS_KEY_ID"),
			config.SourceFile("AWS_PEERING_ACCESS_KEY_ID"),
		),
		config.Path(
			"peering.secretAccessKey",
			config.Sensitive(),
			config.SourceEnv("AWS_PEERING_SECRET_ACCESS_KEY"),
			config.SourceFile("AWS_PEERING_SECRET_ACCESS_KEY"),
		),
		config.Path(
			"peering.assumeRoleName",
			config.SourceEnv("AWS_PEERING_ROLE_NAME"),
			config.SourceFile("AWS_PEERING_ROLE_NAME"),
		),
		config.Path(
			"backupRoleName",
			config.DefaultScalar("CloudManagerBackupServiceRole"),
			config.SourceEnv("AWS_BACKUP_ROLE_NAME"),
		),
		config.Path(
			"efsCapacityCheckInterval",
			config.DefaultScalar(1*time.Hour),
			config.SourceEnv("AWS_EFS_CAPACITY_CHECK_INTERVAL"),
		),
	)

}

// AfterConfigLoaded normalizes the EFS orphan-recovery mode after the config file is decoded.
// Invalid or empty modes silently fall back to "off" (the safe default); there is no logger
// available at this stage, so normalization is quiet by design.
func (c *AwsConfigStruct) AfterConfigLoaded() {
	mode := strings.ToLower(strings.TrimSpace(c.EfsOrphanRecovery.Mode))
	switch mode {
	case EfsOrphanRecoveryModeDryRun, EfsOrphanRecoveryModeOn:
		c.EfsOrphanRecovery.Mode = mode
	default:
		c.EfsOrphanRecovery.Mode = EfsOrphanRecoveryModeOff
	}
}
