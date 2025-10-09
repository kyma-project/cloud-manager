package util

import (
	"testing"

	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/stretchr/testify/assert"
)

func TestArn(t *testing.T) {

	t.Run("StripArnResourceType", func(t *testing.T) {
		for in, expected := range map[string]string{
			"role/roleName": "roleName",
			"role:roleName": "roleName",
			"roleName":      "roleName",
		} {
			res := StripArnResourceType(in)
			assert.Equal(t, expected, res)
		}
	})

	t.Run("ParseArnResourceId", func(t *testing.T) {
		for in, expected := range map[string]string{
			"invalid-arn":                       "",
			"arn:aws:iam::123123:role":          "role",
			"arn:aws:iam::123123:role/":         "",
			"arn:aws:iam::123123:role/roleName": "roleName",
		} {
			res := ParseArnResourceId(in)
			assert.Equal(t, expected, res)
		}
	})

	awsconfig.AwsConfig.Default.AssumeRoleName = "DefaultRole"
	awsconfig.AwsConfig.Peering.AssumeRoleName = "PeeringRole"
	awsconfig.AwsConfig.BackupRoleName = "BackupRole"

	t.Run("partition custom", func(t *testing.T) {
		awsconfig.AwsConfig.ArnPartition = "aws-cn"

		t.Run("RoleArn", func(t *testing.T) {
			x := RoleArn("123123", "roleName")
			assert.Equal(t, "arn:aws-cn:iam::123123:role/roleName", x)
		})

		t.Run("AssumeRoleArnDefault", func(t *testing.T) {
			x := RoleArnDefault("123123")
			assert.Equal(t, "arn:aws-cn:iam::123123:role/DefaultRole", x)
		})

		t.Run("AssumeRoleArnPeering", func(t *testing.T) {
			x := RoleArnPeering("123123")
			assert.Equal(t, "arn:aws-cn:iam::123123:role/PeeringRole", x)
		})

		t.Run("AssumeRoleArnBackup", func(t *testing.T) {
			x := RoleArnBackup("123123")
			assert.Equal(t, "arn:aws-cn:iam::123123:role/BackupRole", x)
		})

		t.Run("EfsArn", func(t *testing.T) {
			x := EfsArn("us-east-1", "123123", "fs-123456")
			assert.Equal(t, "arn:aws-cn:elasticfilesystem:us-east-1:123123:file-system/fs-123456", x)
		})

		t.Run("BackupVaultArn", func(t *testing.T) {
			x := BackupVaultArn("us-east-1", "123123", "my-vault")
			assert.Equal(t, "arn:aws-cn:backup:us-east-1:123123:backup-vault:my-vault", x)
		})

		t.Run("BackupRecoveryPointArn", func(t *testing.T) {
			x := BackupRecoveryPointArn("us-east-1", "123123", "my-point")
			assert.Equal(t, "arn:aws-cn:backup:us-east-1:123123:recovery-point:my-point", x)
		})

	})

	t.Run("partition default", func(t *testing.T) {
		awsconfig.AwsConfig.ArnPartition = "aws"

		t.Run("RoleArn", func(t *testing.T) {
			x := RoleArn("123123", "roleName")
			assert.Equal(t, "arn:aws:iam::123123:role/roleName", x)
		})

		t.Run("AssumeRoleArnDefault", func(t *testing.T) {
			x := RoleArnDefault("123123")
			assert.Equal(t, "arn:aws:iam::123123:role/DefaultRole", x)
		})

		t.Run("AssumeRoleArnPeering", func(t *testing.T) {
			x := RoleArnPeering("123123")
			assert.Equal(t, "arn:aws:iam::123123:role/PeeringRole", x)
		})

		t.Run("AssumeRoleArnBackup", func(t *testing.T) {
			x := RoleArnBackup("123123")
			assert.Equal(t, "arn:aws:iam::123123:role/BackupRole", x)
		})

		t.Run("EfsArn", func(t *testing.T) {
			x := EfsArn("us-east-1", "123123", "fs-123456")
			assert.Equal(t, "arn:aws:elasticfilesystem:us-east-1:123123:file-system/fs-123456", x)
		})

		t.Run("BackupVaultArn", func(t *testing.T) {
			x := BackupVaultArn("us-east-1", "123123", "my-vault")
			assert.Equal(t, "arn:aws:backup:us-east-1:123123:backup-vault:my-vault", x)
		})

		t.Run("BackupRecoveryPointArn", func(t *testing.T) {
			x := BackupRecoveryPointArn("us-east-1", "123123", "my-point")
			assert.Equal(t, "arn:aws:backup:us-east-1:123123:recovery-point:my-point", x)
		})
	})

}
