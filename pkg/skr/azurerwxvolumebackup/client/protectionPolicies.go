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
		Properties: to.Ptr(armrecoveryservicesbackup.AzureFileShareProtectionPolicy{
			BackupManagementType:           to.Ptr("AzureStorage"),
			ProtectedItemsCount:            nil,
			ResourceGuardOperationRequests: nil,
			RetentionPolicy: to.Ptr(armrecoveryservicesbackup.LongTermRetentionPolicy{
				RetentionPolicyType: to.Ptr("LongTermRetentionPolicy"),
				DailySchedule: to.Ptr(armrecoveryservicesbackup.DailyRetentionSchedule{
					RetentionDuration: to.Ptr(armrecoveryservicesbackup.RetentionDuration{
						Count:        to.Ptr(int32(30)),
						DurationType: to.Ptr(armrecoveryservicesbackup.RetentionDurationTypeDays),
					}),
					RetentionTimes: []*time.Time{to.Ptr(time.Now())},
				}),
				MonthlySchedule: nil,
				WeeklySchedule:  nil,
				YearlySchedule:  nil,
			}),
			SchedulePolicy: to.Ptr(armrecoveryservicesbackup.SimpleSchedulePolicy{
				SchedulePolicyType:      nil,
				HourlySchedule:          nil,
				ScheduleRunDays:         nil,
				ScheduleRunFrequency:    to.Ptr(armrecoveryservicesbackup.ScheduleRunTypeDaily),
				ScheduleRunTimes:        []*time.Time{to.Ptr(time.Now())},
				ScheduleWeeklyFrequency: nil,
			}),
			TimeZone:             to.Ptr("UTC"),
			VaultRetentionPolicy: nil,
			WorkLoadType:         to.Ptr(armrecoveryservicesbackup.WorkloadTypeAzureFileShare),
		}),
		Tags: map[string]*string{"cloud-manager": to.Ptr("rwxVolumeBackup")},
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
