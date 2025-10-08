package config

import (
	"time"

	"github.com/kyma-project/cloud-manager/pkg/config"
)

type AwsConfigStruct struct {
	ArnPartition             string        `json:"arnPartition" yaml:"arnPartition"`
	Default                  AwsCreds      `json:"default" yaml:"default"`
	Peering                  AwsCreds      `json:"peering" yaml:"peering"`
	BackupRoleName           string        `json:"backupRoleName" yaml:"backupRoleName"`
	EfsCapacityCheckInterval time.Duration `json:"efsCapacityCheckInterval" yaml:"efsCapacityCheckInterval"`
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
