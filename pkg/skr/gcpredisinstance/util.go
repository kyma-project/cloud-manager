package gcpredisinstance

import (
	"errors"
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

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.PrimaryEndpoint) > 0 {
		result["primaryEndpoint"] = []byte(kcpRedis.Status.PrimaryEndpoint)

		splitEndpoint := strings.Split(kcpRedis.Status.PrimaryEndpoint, ":")
		if len(splitEndpoint) >= 2 {
			host := splitEndpoint[0]
			port := splitEndpoint[1]
			result["host"] = []byte(host)
			result["port"] = []byte(port)
		}
	}

	if len(kcpRedis.Status.ReadEndpoint) > 0 {
		result["readEndpoint"] = []byte(kcpRedis.Status.ReadEndpoint)

		splitReadEndpoint := strings.Split(kcpRedis.Status.ReadEndpoint, ":")
		if len(splitReadEndpoint) >= 2 {
			readHost := splitReadEndpoint[0]
			readPort := splitReadEndpoint[1]
			result["readHost"] = []byte(readHost)
			result["readPort"] = []byte(readPort)
		}
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	if len(kcpRedis.Status.CaCert) > 0 {
		result["CaCert.pem"] = []byte(kcpRedis.Status.CaCert)
	}

	return result
}

func parseAuthSecretExtraData(extraDataTemplates map[string]string, authSecretBaseData map[string][]byte) map[string][]byte {
	baseDataStringMap := map[string]string{}
	for k, v := range authSecretBaseData {
		baseDataStringMap[k] = string(v)
	}

	return util.ParseTemplatesMapToBytesMap(extraDataTemplates, baseDataStringMap)
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

type gcpRedisTierValue struct {
	Tier         string
	MemorySizeGb int32
}

var gcpRedisTierToGcpRedisTierValueMap = map[cloudresourcesv1beta1.GcpRedisTier]gcpRedisTierValue{
	cloudresourcesv1beta1.GcpRedisTierS1: {"BASIC", 1},
	cloudresourcesv1beta1.GcpRedisTierS2: {"BASIC", 3},
	cloudresourcesv1beta1.GcpRedisTierS3: {"BASIC", 6},
	cloudresourcesv1beta1.GcpRedisTierS4: {"BASIC", 12},
	cloudresourcesv1beta1.GcpRedisTierS5: {"BASIC", 24},
	cloudresourcesv1beta1.GcpRedisTierS6: {"BASIC", 48},
	cloudresourcesv1beta1.GcpRedisTierS7: {"BASIC", 101},
	cloudresourcesv1beta1.GcpRedisTierS8: {"BASIC", 200},

	cloudresourcesv1beta1.GcpRedisTierP1: {"STANDARD_HA", 5},
	cloudresourcesv1beta1.GcpRedisTierP2: {"STANDARD_HA", 12},
	cloudresourcesv1beta1.GcpRedisTierP3: {"STANDARD_HA", 24},
	cloudresourcesv1beta1.GcpRedisTierP4: {"STANDARD_HA", 48},
	cloudresourcesv1beta1.GcpRedisTierP5: {"STANDARD_HA", 101},
	cloudresourcesv1beta1.GcpRedisTierP6: {"STANDARD_HA", 200},
}

func redisTierToTierAndMemorySizeConverter(redisTier cloudresourcesv1beta1.GcpRedisTier) (string, int32, error) {
	gcpRedisTierValue, exists := gcpRedisTierToGcpRedisTierValueMap[redisTier]

	if !exists {
		return "", 0, errors.New("unknown redis tier")
	}

	return gcpRedisTierValue.Tier, gcpRedisTierValue.MemorySizeGb, nil
}

func redisTierToReplicaCount(awsRedisTier cloudresourcesv1beta1.GcpRedisTier) int32 {
	if strings.HasPrefix(string(awsRedisTier), "P") {
		return 1
	}
	return 0
}
