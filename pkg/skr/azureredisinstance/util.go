package azureredisinstance

import (
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

func getAuthSecretData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
	result := map[string][]byte{
		"primaryEndpoint": []byte(kcpRedis.Status.PrimaryEndpoint),
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	return result
}
