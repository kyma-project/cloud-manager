package awsredisinstance

import (
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

func getAuthSecretData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
	result := map[string][]byte{
		"primaryEndpoint": []byte(kcpRedis.Status.PrimaryEndpoint),
	}

	if len(kcpRedis.Status.ReadEndpoint) > 0 {
		result["readEndpoint"] = []byte(kcpRedis.Status.ReadEndpoint)
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	return result
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
