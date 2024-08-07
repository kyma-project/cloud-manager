package client

import (
	"fmt"

	redispb "cloud.google.com/go/redis/apiv1/redispb"
	"github.com/elliotchance/pie/v2"
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

func ToMaintenancePolicy(windows *[]cloudcontrolv1beta1.WeeklyMaintenanceWindowGcp) *redispb.MaintenancePolicy {
	if windows == nil {
		return nil
	}

	maintanceWindows := pie.Map(*windows, func(window cloudcontrolv1beta1.WeeklyMaintenanceWindowGcp) *redispb.WeeklyMaintenanceWindow {
		return &redispb.WeeklyMaintenanceWindow{
			Day: dayofweek.DayOfWeek(dayofweek.DayOfWeek_value[window.Day]),
			StartTime: &timeofday.TimeOfDay{
				Hours:   window.StartTime.Hours,
				Minutes: window.StartTime.Minutes,
				Seconds: window.StartTime.Seconds,
				Nanos:   0,
			},
		}
	})

	return &redispb.MaintenancePolicy{
		WeeklyMaintenanceWindow: maintanceWindows,
	}
}
