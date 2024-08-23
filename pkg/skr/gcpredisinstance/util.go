package gcpredisinstance

import (
	"bytes"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(gcpRedis *cloudresourcesv1beta1.GcpRedisInstance) string {
	if gcpRedis.Spec.AuthSecret != nil && len(gcpRedis.Spec.AuthSecret.Name) > 0 {
		return gcpRedis.Spec.AuthSecret.Name
	}

	return gcpRedis.Name
}

func getAuthSecretLabels(gcpRedis *cloudresourcesv1beta1.GcpRedisInstance) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if gcpRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range gcpRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, gcpRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceNamespace, gcpRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(gcpRedis *cloudresourcesv1beta1.GcpRedisInstance) map[string]string {
	if gcpRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range gcpRedis.Spec.AuthSecret.Annotations {
		result[k] = v
	}
	return result
}

func getAuthSecretData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.PrimaryEndpoint) > 0 {
		splitEndpoint := strings.Split(kcpRedis.Status.PrimaryEndpoint, ":")
		host := splitEndpoint[0]
		port := splitEndpoint[1]

		result["primaryEndpoint"] = []byte(kcpRedis.Status.PrimaryEndpoint)
		result["host"] = []byte(host)
		result["port"] = []byte(port)
	}

	if len(kcpRedis.Status.ReadEndpoint) > 0 {
		splitReadEndpoint := strings.Split(kcpRedis.Status.ReadEndpoint, ":")
		readHost := splitReadEndpoint[0]
		readPort := splitReadEndpoint[1]

		result["readEndpoint"] = []byte(kcpRedis.Status.ReadEndpoint)
		result["readHost"] = []byte(readHost)
		result["readPort"] = []byte(readPort)
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	if len(kcpRedis.Status.CaCert) > 0 {
		result["CaCert.pem"] = []byte(kcpRedis.Status.CaCert)
	}

	return result
}

func toGcpMaintenancePolicy(maintenancePolicy *cloudresourcesv1beta1.MaintenancePolicy) *cloudcontrolv1beta1.MaintenancePolicyGcp {
	if maintenancePolicy == nil {
		return nil
	}

	if maintenancePolicy.DayOfWeek == nil {
		return nil
	}

	return &cloudcontrolv1beta1.MaintenancePolicyGcp{
		DayOfWeek: &cloudcontrolv1beta1.DayOfWeekPolicyGcp{
			Day: maintenancePolicy.DayOfWeek.Day,
			StartTime: cloudcontrolv1beta1.TimeOfDayGcp{
				Hours:   maintenancePolicy.DayOfWeek.StartTime.Hours,
				Minutes: maintenancePolicy.DayOfWeek.StartTime.Minutes,
			},
		},
	}
}

func toGcpTransitEncryption(transitEncryption *cloudresourcesv1beta1.TransitEncryption) *cloudcontrolv1beta1.TransitEncryptionGcp {
	if transitEncryption == nil {
		return nil
	}

	return &cloudcontrolv1beta1.TransitEncryptionGcp{
		ServerAuthentication: transitEncryption.ServerAuthentication,
	}
}

func areMapsDifferent(first, second map[string]string) bool {
	if len(first) != len(second) {
		return true
	}

	for key, value1 := range first {
		value2, exists := second[key]
		if !exists || value1 != value2 {
			return true
		}
	}

	return false
}

func areByteMapsEqual(first, second map[string][]byte) bool {
	if len(first) != len(second) {
		return false
	}

	for key, firstValue := range first {
		secondValue, exists := second[key]
		if !exists {
			return false
		}

		if !bytes.Equal(firstValue, secondValue) {
			return false
		}
	}

	return true
}
