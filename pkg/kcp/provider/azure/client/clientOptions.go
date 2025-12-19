package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
)

type CredentialOptionsBuilder struct {
	options *azidentity.ClientSecretCredentialOptions
}

func NewCredentialOptions() *CredentialOptionsBuilder {
	b := &CredentialOptionsBuilder{}
	if config.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		b.options = &azidentity.ClientSecretCredentialOptions{}
		b.options.Cloud = cloud.AzureChina
	}
	return b
}
func (b *CredentialOptionsBuilder) WithAnyTenant() *CredentialOptionsBuilder {
	if b.options == nil {
		b.options = &azidentity.ClientSecretCredentialOptions{}
	}
	b.options.AdditionallyAllowedTenants = []string{"*"}
	return b
}

func (b *CredentialOptionsBuilder) Build() *azidentity.ClientSecretCredentialOptions {
	return b.options
}

type OptionsBuilder struct {
	options *arm.ClientOptions
}

func NewClientOptions() *OptionsBuilder {
	b := &OptionsBuilder{}
	if config.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		b.options = &arm.ClientOptions{}
		b.options.Cloud = cloud.AzureChina
	}
	return b
}

func (b *OptionsBuilder) WithAuxiliaryTenants(auxiliaryTenants []string) *OptionsBuilder {
	if b.options == nil {
		b.options = &arm.ClientOptions{}
	}
	b.options.AuxiliaryTenants = auxiliaryTenants
	return b
}

func (b *OptionsBuilder) Build() *arm.ClientOptions {
	return b.options
}
