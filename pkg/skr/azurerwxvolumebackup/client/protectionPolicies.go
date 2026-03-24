package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"time"
)

type ProtectionPoliciesClient interface {
	CreateBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error
	DeleteBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error
}

type protectionPoliciesClient struct {
	azureClient *armrecoveryservicesbackup.ProtectionPoliciesClient
}

func NewProtectionPoliciesClient(ppc *armrecoveryservicesbackup.ProtectionPoliciesClient) ProtectionPoliciesClient {
	return protectionPoliciesClient{ppc}
}

func (c protectionPoliciesClient) CreateBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {

	ppr := armrecoveryservicesbackup.ProtectionPolicyResource{
		ETag:     nil,
		Location: nil,
		Properties: new(armrecoveryservicesbackup.AzureFileShareProtectionPolicy{
			BackupManagementType:           new("AzureStorage"),
			ProtectedItemsCount:            nil,
			ResourceGuardOperationRequests: nil,
			RetentionPolicy: new(armrecoveryservicesbackup.LongTermRetentionPolicy{
				RetentionPolicyType: new("LongTermRetentionPolicy"),
				DailySchedule: new(armrecoveryservicesbackup.DailyRetentionSchedule{
					RetentionDuration: new(armrecoveryservicesbackup.RetentionDuration{
						Count:        new(int32(30)),
						DurationType: to.Ptr(armrecoveryservicesbackup.RetentionDurationTypeDays),
					}),
					RetentionTimes: []*time.Time{new(time.Now())},
				}),
				MonthlySchedule: nil,
				WeeklySchedule:  nil,
				YearlySchedule:  nil,
			}),
			SchedulePolicy: new(armrecoveryservicesbackup.SimpleSchedulePolicy{
				SchedulePolicyType:      nil,
				HourlySchedule:          nil,
				ScheduleRunDays:         nil,
				ScheduleRunFrequency:    to.Ptr(armrecoveryservicesbackup.ScheduleRunTypeDaily),
				ScheduleRunTimes:        []*time.Time{new(time.Now())},
				ScheduleWeeklyFrequency: nil,
			}),
			TimeZone:             new("UTC"),
			VaultRetentionPolicy: nil,
			WorkLoadType:         to.Ptr(armrecoveryservicesbackup.WorkloadTypeAzureFileShare),
		}),
		Tags: map[string]*string{"cloud-manager": new("rwxVolumeBackup")},
		ID:   nil,
		Name: nil,
		Type: nil,
	}

	_, err := c.azureClient.CreateOrUpdate(ctx, vaultName, resourceGroupName, policyName, ppr, nil)
	if err != nil {
		return err
	}
	return nil

}

// poller doesn't return a response body
func (c protectionPoliciesClient) DeleteBackupPolicy(ctx context.Context, vaultName string, resourceGroupName string, policyName string) error {

	_, err := c.azureClient.BeginDelete(ctx, vaultName, resourceGroupName, policyName, nil)
	if err != nil {
		return err
	}

	return nil
}
