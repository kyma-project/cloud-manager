package azureredisinstance

import (
	"errors"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(azureRedis *cloudresourcesv1beta1.AzureRedisInstance) string {
	if azureRedis.Spec.AuthSecret != nil && len(azureRedis.Spec.AuthSecret.Name) > 0 {
		return azureRedis.Spec.AuthSecret.Name
	}

	return azureRedis.Name
}

func getAuthSecretLabels(azureRedis *cloudresourcesv1beta1.AzureRedisInstance) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if azureRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range azureRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, azureRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceNamespace, azureRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(azureRedis *cloudresourcesv1beta1.AzureRedisInstance) map[string]string {
	if azureRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range azureRedis.Spec.AuthSecret.Annotations {
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

	return result
}

func parseAuthSecretExtraData(extraDataTemplates map[string]string, authSecretBaseData map[string][]byte) map[string][]byte {
	baseDataStringMap := map[string]string{}
	for k, v := range authSecretBaseData {
		baseDataStringMap[k] = string(v)
	}

	return util.ParseTemplatesMapToBytesMap(extraDataTemplates, baseDataStringMap)
}

type azureRedisTierValue struct {
	Tier     string
	Capacity int
}

var azureRedisTierToAzureRedisSKUCapacityValueMap = map[cloudresourcesv1beta1.AzureRedisTier]azureRedisTierValue{
	cloudresourcesv1beta1.AzureRedisTierP1: {"P", 1},
	cloudresourcesv1beta1.AzureRedisTierP2: {"P", 2},
	cloudresourcesv1beta1.AzureRedisTierP3: {"P", 3},
	cloudresourcesv1beta1.AzureRedisTierP4: {"P", 4},
	cloudresourcesv1beta1.AzureRedisTierP5: {"P", 5},

	cloudresourcesv1beta1.AzureRedisTierS1: {"S", 1},
	cloudresourcesv1beta1.AzureRedisTierS2: {"S", 2},
	cloudresourcesv1beta1.AzureRedisTierS3: {"S", 3},
	cloudresourcesv1beta1.AzureRedisTierS4: {"S", 4},
	cloudresourcesv1beta1.AzureRedisTierS5: {"S", 5},
}

func RedisTierToSKUCapacityConverter(redisTier cloudresourcesv1beta1.AzureRedisTier) (string, int, error) {
	azureRedisSKUValue, exists := azureRedisTierToAzureRedisSKUCapacityValueMap[redisTier]

	if !exists {
		return "", 0, errors.New("unknown azure redis tier")
	}

	return azureRedisSKUValue.Tier, azureRedisSKUValue.Capacity, nil
}
