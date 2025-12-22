package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClientOptions(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	o := NewClientOptionsBuilder().Build()
	assert.Equal(t, (*arm.ClientOptions)(nil), o)
}

func TestNewClientOptionsWithAuxiliaryTenants(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	auxiliaryTenants := []string{"tenant1", "tenant2"}
	o := NewClientOptionsBuilder().WithAuxiliaryTenants(auxiliaryTenants).Build()
	assert.Equal(t, auxiliaryTenants, o.AuxiliaryTenants)
}

func TestNewClientOptionsChina(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{
		ClientOptions: config.ClientOptions{
			Cloud: "AzureChina",
		},
	}
	o := NewClientOptionsBuilder().Build()
	assert.Equal(t, cloud.AzureChina, o.Cloud)
}

func TestNewClientOptionsWithAuxiliaryTenantsChina(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{
		ClientOptions: config.ClientOptions{
			Cloud: "AzureChina",
		},
	}
	auxiliaryTenants := []string{"tenant1", "tenant2"}
	o := NewClientOptionsBuilder().WithAuxiliaryTenants(auxiliaryTenants).Build()
	assert.Equal(t, auxiliaryTenants, o.AuxiliaryTenants)
	assert.Equal(t, cloud.AzureChina, o.Cloud)
}

func TestNewCredentialOptions(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	o := NewCredentialOptionsBuilder().Build()
	assert.Equal(t, (*azidentity.ClientSecretCredentialOptions)(nil), o)
}

func TestNewCredentialOptionsWithAnyTenant(t *testing.T) {
	o := NewCredentialOptionsBuilder().WithAnyTenant().Build()
	assert.Equal(t, []string{"*"}, o.AdditionallyAllowedTenants)
}

func TestNewCredentialOptionsChina(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{
		ClientOptions: config.ClientOptions{
			Cloud: "AzureChina",
		},
	}
	o := NewClientOptionsBuilder().Build()
	assert.Equal(t, cloud.AzureChina, o.Cloud)
}

func TestNewCredentialOptionsWithAnyTenantChina(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{
		ClientOptions: config.ClientOptions{
			Cloud: "AzureChina",
		},
	}
	o := NewCredentialOptionsBuilder().WithAnyTenant().Build()
	assert.Equal(t, []string{"*"}, o.AdditionallyAllowedTenants)
	assert.Equal(t, cloud.AzureChina, o.Cloud)
}
