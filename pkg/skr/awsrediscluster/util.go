package awsrediscluster

import (
	"errors"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(awsRedis *cloudresourcesv1beta1.AwsRedisCluster) string {
	if awsRedis.Spec.AuthSecret != nil && len(awsRedis.Spec.AuthSecret.Name) > 0 {
		return awsRedis.Spec.AuthSecret.Name
	}

	return awsRedis.Name
}

func getAuthSecretLabels(awsRedis *cloudresourcesv1beta1.AwsRedisCluster) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if awsRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range awsRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterStatusId, awsRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisClusterNamespace, awsRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(awsRedis *cloudresourcesv1beta1.AwsRedisCluster) map[string]string {
	if awsRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range awsRedis.Spec.AuthSecret.Annotations {
		result[k] = v
	}
	return result
}

func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisCluster) map[string][]byte {
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

	return result
}

func parseAuthSecretExtraData(extraDataTemplates map[string]string, authSecretBaseData map[string][]byte) map[string][]byte {
	baseDataStringMap := map[string]string{}
	for k, v := range authSecretBaseData {
		baseDataStringMap[k] = string(v)
	}

	return util.ParseTemplatesMapToBytesMap(extraDataTemplates, baseDataStringMap)
}

var AwsRedisClusterTierToCacheNodeTypeMap = map[cloudresourcesv1beta1.AwsRedisClusterTier]string{
	cloudresourcesv1beta1.AwsRedisTierC1: "cache.t4g.small",
	cloudresourcesv1beta1.AwsRedisTierC2: "cache.t4g.medium",
	cloudresourcesv1beta1.AwsRedisTierC3: "cache.m7g.large",
	cloudresourcesv1beta1.AwsRedisTierC4: "cache.m7g.xlarge",
	cloudresourcesv1beta1.AwsRedisTierC5: "cache.m7g.2xlarge",
	cloudresourcesv1beta1.AwsRedisTierC6: "cache.m7g.4xlarge",
	cloudresourcesv1beta1.AwsRedisTierC7: "cache.m7g.8xlarge",
	cloudresourcesv1beta1.AwsRedisTierC8: "cache.m7g.16xlarge",
}

func redisTierToCacheNodeTypeConvertor(AwsRedisClusterTier cloudresourcesv1beta1.AwsRedisClusterTier, fromConfigOverwrite map[cloudresourcesv1beta1.AwsRedisClusterTier]string) (string, error) {
	if fromConfigOverwrite != nil {
		if overrideNode, exists := fromConfigOverwrite[AwsRedisClusterTier]; exists {
			return overrideNode, nil
		}
	}

	cacheNode, exists := AwsRedisClusterTierToCacheNodeTypeMap[AwsRedisClusterTier]

	if !exists {
		return "", errors.New("unknown redis tier")
	}

	return cacheNode, nil
}
