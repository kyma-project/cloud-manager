package gcprediscluster

import (
	"errors"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(gcpRedis *cloudresourcesv1beta1.GcpRedisCluster) string {
	if gcpRedis.Spec.AuthSecret != nil && len(gcpRedis.Spec.AuthSecret.Name) > 0 {
		return gcpRedis.Spec.AuthSecret.Name
	}

	return gcpRedis.Name
}

func getAuthSecretLabels(gcpRedis *cloudresourcesv1beta1.GcpRedisCluster) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if gcpRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range gcpRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterStatusId, gcpRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterNamespace, gcpRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(gcpRedis *cloudresourcesv1beta1.GcpRedisCluster) map[string]string {
	if gcpRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range gcpRedis.Spec.AuthSecret.Annotations {
		result[k] = v
	}
	return result
}

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.GcpRedisCluster) map[string][]byte {
	result := map[string][]byte{}

	if len(kcpRedis.Status.DiscoveryEndpoint) > 0 {
		result["discoveryEndpoint"] = []byte(kcpRedis.Status.DiscoveryEndpoint)

		splitEndpoint := strings.Split(kcpRedis.Status.DiscoveryEndpoint, ":")
		if len(splitEndpoint) >= 2 {
			host := splitEndpoint[0]
			port := splitEndpoint[1]
			result["host"] = []byte(host)
			result["port"] = []byte(port)
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

var gcpRedisTierToGcpRedisTierValueMap = map[cloudresourcesv1beta1.GcpRedisClusterTier]string{
	cloudresourcesv1beta1.GcpRedisClusterTierC1: "REDIS_SHARED_CORE_NANO",
	cloudresourcesv1beta1.GcpRedisClusterTierC3: "REDIS_STANDARD_SMALL",
	cloudresourcesv1beta1.GcpRedisClusterTierC4: "REDIS_HIGHMEM_MEDIUM",
	cloudresourcesv1beta1.GcpRedisClusterTierC6: "REDIS_HIGHMEM_XLARGE",
}

func redisTierToNodeTypeConverter(redisTier cloudresourcesv1beta1.GcpRedisClusterTier) (string, error) {
	gcpRedisTier, exists := gcpRedisTierToGcpRedisTierValueMap[redisTier]

	if !exists {
		return "", errors.New("unknown redis tier")
	}

	return gcpRedisTier, nil
}
