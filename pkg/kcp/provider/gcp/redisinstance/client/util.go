package client

import (
	"fmt"

	redispb "cloud.google.com/go/redis/apiv1/redispb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
)

func GetGcpMemoryStoreRedisName(projectId, locationId, instanceId string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, locationId, GetGcpMemoryStoreRedisInstanceId(instanceId))
}

func GetGcpMemoryStoreRedisInstanceId(instanceId string) string {
	return fmt.Sprintf("cm-%s", instanceId)
}

func ToMaintenancePolicy(maintenancePolicy *cloudcontrolv1beta1.MaintenancePolicyGcp) *redispb.MaintenancePolicy {
	if maintenancePolicy == nil {
		return nil
	}

	if maintenancePolicy.DayOfWeek == nil {
		return nil
	}

	maintenanceWindow := &redispb.WeeklyMaintenanceWindow{
		Day: dayofweek.DayOfWeek(dayofweek.DayOfWeek_value[maintenancePolicy.DayOfWeek.Day]),
		StartTime: &timeofday.TimeOfDay{
			Hours:   maintenancePolicy.DayOfWeek.StartTime.Hours,
			Minutes: maintenancePolicy.DayOfWeek.StartTime.Minutes,
			Seconds: 0,
			Nanos:   0,
		},
	}

	return &redispb.MaintenancePolicy{
		WeeklyMaintenanceWindow: []*redispb.WeeklyMaintenanceWindow{maintenanceWindow},
	}
}

func ToTransitEncryptionMode(transitEncryption *cloudcontrolv1beta1.TransitEncryptionGcp) redispb.Instance_TransitEncryptionMode {
	if transitEncryption == nil {
		return redispb.Instance_DISABLED
	}

	if transitEncryption.ServerAuthentication {
		return redispb.Instance_SERVER_AUTHENTICATION
	}

	return redispb.Instance_DISABLED
}
