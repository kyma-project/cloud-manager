package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
)

func NewCredentialOptions() *azidentity.ClientSecretCredentialOptions {
	credentialOptions := &azidentity.ClientSecretCredentialOptions{}

	if azureconfig.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		credentialOptions.Cloud = cloud.AzureChina
	}

	return credentialOptions
}

func NewClientOptions() *arm.ClientOptions {
	options := &arm.ClientOptions{}

	if azureconfig.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		options.Cloud = cloud.AzureChina
	}

	return options
}
