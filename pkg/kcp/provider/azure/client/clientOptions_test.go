package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremetrics "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/metrics"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewClientOptions(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	o := NewClientOptionsBuilder().Build()
	assert.NotNil(t, o)
	assert.Equal(t, cloud.Configuration{}, o.Cloud)
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, o.PerRetryPolicies)
}

func TestNewClientOptionsWithAuxiliaryTenants(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	auxiliaryTenants := []string{"tenant1", "tenant2"}
	o := NewClientOptionsBuilder().WithAuxiliaryTenants(auxiliaryTenants).Build()
	assert.Equal(t, auxiliaryTenants, o.AuxiliaryTenants)
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, o.PerRetryPolicies)
}

func TestNewClientOptionsChina(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{
		ClientOptions: config.ClientOptions{
			Cloud: "AzureChina",
		},
	}
	o := NewClientOptionsBuilder().Build()
	assert.Equal(t, cloud.AzureChina, o.Cloud)
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, o.PerRetryPolicies)
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
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, o.PerRetryPolicies)
}

func TestNewClientOptionsBuildIsIdempotent(t *testing.T) {
	config.AzureConfig = &config.AzureConfigStruct{}
	b := NewClientOptionsBuilder()
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, b.Build().PerRetryPolicies)
	assert.Equal(t, []policy.Policy{azuremetrics.NewMetricsPolicy()}, b.Build().PerRetryPolicies)
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

type fakeTransport struct {
	statusCode int
	body       string
}

func (t *fakeTransport) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: t.statusCode,
		Status:     fmt.Sprintf("%d %s", t.statusCode, http.StatusText(t.statusCode)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Request:    req,
	}, nil
}

type fakeTokenCredential struct{}

func (fakeTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

func TestBuiltOptionsEmitMetricsFromArmClient(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()
	config.AzureConfig = &config.AzureConfigStruct{}

	opts := NewClientOptionsBuilder().Build()
	opts.Transport = &fakeTransport{statusCode: http.StatusOK, body: "{}"}
	opts.Retry = policy.RetryOptions{MaxRetries: -1}

	factory, err := armnetwork.NewClientFactory("test-subscription-id", fakeTokenCredential{}, opts)
	assert.NoError(t, err)

	_, err = factory.NewVirtualNetworksClient().Get(context.Background(), "rg-1", "vnet-1", nil)
	assert.NoError(t, err)

	value := testutil.ToFloat64(metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderAzure,
		"GET /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Network/virtualNetworks/{id}",
		"200",
		"",
		"test-subscription-id",
	))
	assert.Equal(t, float64(1), value)
}
