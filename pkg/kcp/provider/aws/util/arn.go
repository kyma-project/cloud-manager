package util

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
)

var arnResourceTypeRegex = regexp.MustCompile("[:/]")

func StripArnResourceType(resource string) string {
	arr := arnResourceTypeRegex.Split(resource, -1)
	if len(arr) == 1 {
		return arr[0]
	}
	return arr[1]
}

func ParseArnResourceId(arnStr string) string {
	a, err := arn.Parse(arnStr)
	if err != nil {
		return ""
	}
	return StripArnResourceType(a.Resource)
}

// IAM Role =====================================================================

// RoleArn returns string arn in form "arn:<partition>:iam::<accountId>:role/<roleName>"
func RoleArn(accountId, roleName string) string {
	a := &arn.ARN{
		Partition: awsconfig.AwsConfig.ArnPartition,
		Service:   "iam",
		AccountID: accountId,
		Resource:  fmt.Sprintf("role/%s", roleName),
	}
	return a.String()
}

func RoleArnDefault(accountId string) string {
	return RoleArn(accountId, awsconfig.AwsConfig.Default.AssumeRoleName)
}

func RoleArnPeering(accountId string) string {
	return RoleArn(accountId, awsconfig.AwsConfig.Peering.AssumeRoleName)
}

func RoleArnBackup(accountId string) string {
	return RoleArn(accountId, awsconfig.AwsConfig.BackupRoleName)
}

// EFS ============================================================================

// EfsArn returns string arn in form "arn:<partition>:elasticfilesystem:<region>:<accountId>:file-system/<fileSystemId>"
func EfsArn(region, accountId, fileSystemId string) string {
	a := &arn.ARN{
		Partition: awsconfig.AwsConfig.ArnPartition,
		Service:   "elasticfilesystem",
		Region:    region,
		AccountID: accountId,
		Resource:  fmt.Sprintf("file-system/%s", fileSystemId),
	}
	return a.String()
}

// Backup =====================================================================

// BackupVaultArn returns string arn in form "arn:<partition>:backup:<region>:<accountId>:backup-vault:<vaultName>"
func BackupVaultArn(region, accountId, vaultName string) string {
	a := &arn.ARN{
		Partition: awsconfig.AwsConfig.ArnPartition,
		Service:   "backup",
		Region:    region,
		AccountID: accountId,
		Resource:  fmt.Sprintf("backup-vault:%s", vaultName),
	}
	return a.String()
}

// BackupRecoveryPointArn returns string arn in form "arn:<partition>:backup:<region>:<accountId>:recovery-point:<recoveryPointId>"
func BackupRecoveryPointArn(region, accountId, recoveryPointId string) string {
	a := &arn.ARN{
		Partition: awsconfig.AwsConfig.ArnPartition,
		Service:   "backup",
		Region:    region,
		AccountID: accountId,
		Resource:  fmt.Sprintf("recovery-point:%s", recoveryPointId),
	}
	return a.String()
}
