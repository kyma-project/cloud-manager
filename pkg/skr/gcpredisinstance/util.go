package gcpredisinstance

import (
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
	result := map[string][]byte{
		"primaryEndpoint": []byte(kcpRedis.Status.PrimaryEndpoint),
	}

	if len(kcpRedis.Status.ReadEndpoint) > 0 {
		result["readEndpoint"] = []byte(kcpRedis.Status.ReadEndpoint)
	}

	if len(kcpRedis.Status.AuthString) > 0 {
		result["authString"] = []byte(kcpRedis.Status.AuthString)
	}

	if len(kcpRedis.Status.CaCert) > 0 {
		result["CaCert.pem"] = []byte(kcpRedis.Status.CaCert)
	}

	return result
}
