package lib

import cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"

const AliasLabel = "e2e.kyma-project.io/alias"

var DefaultRegions = map[cloudcontrolv1beta1.ProviderType]string{
	cloudcontrolv1beta1.ProviderAws:       "us-east-1",
	cloudcontrolv1beta1.ProviderGCP:       "us-east1",
	cloudcontrolv1beta1.ProviderAzure:     "westeurope",
	cloudcontrolv1beta1.ProviderOpenStack: "eu-de-1",
}

const (
	ExpiresAtAnnotation               = "operator.kyma-project.io/expires-at"
	ForceKubeconfigRotationAnnotation = "operator.kyma-project.io/force-kubeconfig-rotation"
)
