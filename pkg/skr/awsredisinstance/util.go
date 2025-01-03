package awsredisinstance

import (
	"errors"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func getAuthSecretName(awsRedis *cloudresourcesv1beta1.AwsRedisInstance) string {
	if awsRedis.Spec.AuthSecret != nil && len(awsRedis.Spec.AuthSecret.Name) > 0 {
		return awsRedis.Spec.AuthSecret.Name
	}

	return awsRedis.Name
}

func getAuthSecretLabels(awsRedis *cloudresourcesv1beta1.AwsRedisInstance) map[string]string {
	labelsBuilder := util.NewLabelBuilder()

	if awsRedis.Spec.AuthSecret != nil {
		for labelName, labelValue := range awsRedis.Spec.AuthSecret.Labels {
			labelsBuilder.WithCustomLabel(labelName, labelValue)
		}
	}

	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, awsRedis.Status.Id)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelRedisInstanceNamespace, awsRedis.Namespace)
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelCloudManaged, "true")
	labelsBuilder.WithCloudManagerDefaults()
	pvLabels := labelsBuilder.Build()

	return pvLabels
}

func getAuthSecretAnnotations(awsRedis *cloudresourcesv1beta1.AwsRedisInstance) map[string]string {
	if awsRedis.Spec.AuthSecret == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range awsRedis.Spec.AuthSecret.Annotations {
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

var awsRedisTierToCacheNodeTypeMap = map[cloudresourcesv1beta1.AwsRedisTier]string{
	cloudresourcesv1beta1.AwsRedisTierS1: "cache.t4g.small",
	cloudresourcesv1beta1.AwsRedisTierS2: "cache.t4g.medium",
	cloudresourcesv1beta1.AwsRedisTierS3: "cache.m7g.large",
	cloudresourcesv1beta1.AwsRedisTierS4: "cache.m7g.xlarge",
	cloudresourcesv1beta1.AwsRedisTierS5: "cache.m7g.2xlarge",
	cloudresourcesv1beta1.AwsRedisTierS6: "cache.m7g.4xlarge",
	cloudresourcesv1beta1.AwsRedisTierS7: "cache.m7g.8xlarge",
	cloudresourcesv1beta1.AwsRedisTierS8: "cache.m7g.16xlarge",

	cloudresourcesv1beta1.AwsRedisTierP1: "cache.m7g.large",
	cloudresourcesv1beta1.AwsRedisTierP2: "cache.m7g.xlarge",
	cloudresourcesv1beta1.AwsRedisTierP3: "cache.m7g.2xlarge",
	cloudresourcesv1beta1.AwsRedisTierP4: "cache.m7g.4xlarge",
	cloudresourcesv1beta1.AwsRedisTierP5: "cache.m7g.8xlarge",
	cloudresourcesv1beta1.AwsRedisTierP6: "cache.m7g.16xlarge",
}

func redisTierToCacheNodeTypeConvertor(awsRedisTier cloudresourcesv1beta1.AwsRedisTier) (string, error) {
	cacheNode, exists := awsRedisTierToCacheNodeTypeMap[awsRedisTier]

	if !exists {
		return "", errors.New("unknown redis tier")
	}

	return cacheNode, nil
}

func redisTierToReadReplicas(awsRedisTier cloudresourcesv1beta1.AwsRedisTier) int32 {
	if strings.HasPrefix(string(awsRedisTier), "P") {
		return 1
	}
	return 0
}
