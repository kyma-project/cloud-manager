package metrics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestParseAzureRequestPath(t *testing.T) {
	testCases := []struct {
		name                 string
		path                 string
		expectedPath         string
		expectedRegion       string
		expectedSubscription string
	}{
		{
			name:                 "resource with nested child",
			path:                 "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Network/virtualNetworks/vnet-1/virtualNetworkPeerings/peering-1",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Network/virtualNetworks/{id}/virtualNetworkPeerings/{id}",
			expectedSubscription: "sub-1",
		},
		{
			name:                 "resource group with lowercase keyword casing",
			path:                 "/subscriptions/sub-2/resourcegroups/rg-2",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}",
			expectedSubscription: "sub-2",
		},
		{
			name:                 "collection list under resource group",
			path:                 "/subscriptions/sub-3/resourceGroups/rg-3/providers/Microsoft.Cache/redis",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Cache/redis",
			expectedSubscription: "sub-3",
		},
		{
			name:                 "action segment after resource name",
			path:                 "/subscriptions/sub-4/resourceGroups/rg-4/providers/Microsoft.Cache/redis/redis-1/listKeys",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Cache/redis/{id}/listKeys",
			expectedSubscription: "sub-4",
		},
		{
			name:                 "lro polling url with location",
			path:                 "/subscriptions/sub-5/providers/Microsoft.Cache/locations/westeurope/asyncOperations/op-1",
			expectedPath:         "/subscriptions/{id}/providers/Microsoft.Cache/locations/{id}/asyncOperations/{id}",
			expectedRegion:       "westeurope",
			expectedSubscription: "sub-5",
		},
		{
			name:                 "lro operations url with location",
			path:                 "/subscriptions/sub-6/providers/Microsoft.Network/locations/eastus2/operations/op-2",
			expectedPath:         "/subscriptions/{id}/providers/Microsoft.Network/locations/{id}/operations/{id}",
			expectedRegion:       "eastus2",
			expectedSubscription: "sub-6",
		},
		{
			name:         "provider level operations without subscription",
			path:         "/providers/Microsoft.Cache/operations",
			expectedPath: "/providers/Microsoft.Cache/operations",
		},
		{
			name:                 "subscription level action",
			path:                 "/subscriptions/sub-7/providers/Microsoft.Cache/CheckNameAvailability",
			expectedPath:         "/subscriptions/{id}/providers/Microsoft.Cache/CheckNameAvailability",
			expectedSubscription: "sub-7",
		},
		{
			name:                 "extension resource with nested provider",
			path:                 "/subscriptions/sub-8/resourceGroups/rg-8/providers/Microsoft.Storage/storageAccounts/acc-1/providers/Microsoft.Authorization/roleAssignments/ra-1",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Storage/storageAccounts/{id}/providers/Microsoft.Authorization/roleAssignments/{id}",
			expectedSubscription: "sub-8",
		},
		{
			name:                 "uuid on a type position is masked",
			path:                 "/subscriptions/sub-9/resourceGroups/rg-9/providers/Microsoft.RecoveryServices/vaults/vault-1/backupJobs/operationResults/b0d2d95f-9f47-4b3f-a022-56ab0e0a3f9d",
			expectedPath:         "/subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.RecoveryServices/vaults/{id}/backupJobs/{id}/{id}",
			expectedSubscription: "sub-9",
		},
		{
			name:         "empty path",
			path:         "",
			expectedPath: "/",
		},
		{
			name:         "root path",
			path:         "/",
			expectedPath: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitizedPath, region, subscription := parseAzureRequestPath(tc.path)
			assert.Equal(t, tc.expectedPath, sanitizedPath)
			assert.Equal(t, tc.expectedRegion, region)
			assert.Equal(t, tc.expectedSubscription, subscription)
		})
	}
}

type fakeTransport struct {
	statusCode int
	body       string
	err        error
}

func (t *fakeTransport) Do(req *http.Request) (*http.Response, error) {
	if t.err != nil {
		return nil, t.err
	}
	return &http.Response{
		StatusCode: t.statusCode,
		Status:     fmt.Sprintf("%d %s", t.statusCode, http.StatusText(t.statusCode)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Request:    req,
	}, nil
}

// fakeSequenceTransport responds with the given status codes in order, repeating
// the last one once the sequence is exhausted, and records how many attempts the
// pipeline actually made.
type fakeSequenceTransport struct {
	statusCodes []int
	attempts    int
}

func (t *fakeSequenceTransport) Do(req *http.Request) (*http.Response, error) {
	statusCode := t.statusCodes[len(t.statusCodes)-1]
	if t.attempts < len(t.statusCodes) {
		statusCode = t.statusCodes[t.attempts]
	}
	t.attempts++
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader("{}")),
		Request:    req,
	}, nil
}

// newMetricsTestPipeline builds a pipeline with retries disabled, so the tests
// using it observe exactly one attempt per call. TestMetricsPolicyCountsEveryAttempt
// covers the retrying pipeline.
func newMetricsTestPipeline(transport *fakeTransport) runtime.Pipeline {
	return runtime.NewPipeline("metricstest", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
		Transport:        transport,
		Retry:            policy.RetryOptions{MaxRetries: -1},
		PerRetryPolicies: []policy.Policy{NewMetricsPolicy()},
	})
}

func metricValue(method, responseCode, region, subscription string) float64 {
	return testutil.ToFloat64(metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderAzure,
		method,
		responseCode,
		region,
		subscription,
	))
}

func TestMetricsPolicySuccessfulCall(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	pl := newMetricsTestPipeline(&fakeTransport{statusCode: http.StatusOK, body: "{}"})
	req, err := runtime.NewRequest(context.Background(), http.MethodGet,
		"https://management.azure.com/subscriptions/sub-ok/resourceGroups/rg-1/providers/Microsoft.Network/virtualNetworks/vnet-1?api-version=2024-01-01")
	assert.NoError(t, err)

	resp, err := pl.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, float64(1), metricValue(
		"GET /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Network/virtualNetworks/{id}",
		"200", "", "sub-ok"))
}

func TestMetricsPolicyFailedCall(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	pl := newMetricsTestPipeline(&fakeTransport{statusCode: http.StatusNotFound, body: `{"error":{"code":"ResourceNotFound"}}`})
	req, err := runtime.NewRequest(context.Background(), http.MethodDelete,
		"https://management.azure.com/subscriptions/sub-missing/resourceGroups/rg-1/providers/Microsoft.Cache/redis/redis-1")
	assert.NoError(t, err)

	resp, err := pl.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	assert.Equal(t, float64(1), metricValue(
		"DELETE /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Cache/redis/{id}",
		"404", "", "sub-missing"))
}

func TestMetricsPolicyThrottledCall(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	pl := newMetricsTestPipeline(&fakeTransport{statusCode: http.StatusTooManyRequests, body: "{}"})
	req, err := runtime.NewRequest(context.Background(), http.MethodPut,
		"https://management.azure.com/subscriptions/sub-throttled/resourceGroups/rg-1/providers/Microsoft.Network/virtualNetworks/vnet-1")
	assert.NoError(t, err)

	resp, err := pl.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	assert.Equal(t, float64(1), metricValue(
		"PUT /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Network/virtualNetworks/{id}",
		"429", "", "sub-throttled"))
}

// TestMetricsPolicyCountsEveryAttempt pins the reason the policy is registered as
// a per-retry rather than a per-call policy: the 429s the SDK retry policy absorbs
// must still show up in the metric. Registered per-call this records a single 200.
func TestMetricsPolicyCountsEveryAttempt(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	transport := &fakeSequenceTransport{statusCodes: []int{
		http.StatusTooManyRequests,
		http.StatusTooManyRequests,
		http.StatusOK,
	}}
	// MaxRetries is left at zero so azcore applies its production default of 3
	// retries; only the delays are shortened to keep the test fast.
	pl := runtime.NewPipeline("metricstest", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
		Transport:        transport,
		Retry:            policy.RetryOptions{RetryDelay: time.Millisecond, MaxRetryDelay: 10 * time.Millisecond},
		PerRetryPolicies: []policy.Policy{NewMetricsPolicy()},
	})

	req, err := runtime.NewRequest(context.Background(), http.MethodPut,
		"https://management.azure.com/subscriptions/sub-retried/resourceGroups/rg-1/providers/Microsoft.Cache/redis/redis-1")
	assert.NoError(t, err)

	resp, err := pl.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, transport.attempts)

	method := "PUT /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Cache/redis/{id}"
	assert.Equal(t, float64(2), metricValue(method, "429", "", "sub-retried"))
	assert.Equal(t, float64(1), metricValue(method, "200", "", "sub-retried"))
}

func TestMetricsPolicyTransportError(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	pl := newMetricsTestPipeline(&fakeTransport{err: errors.New("connection refused")})
	req, err := runtime.NewRequest(context.Background(), http.MethodGet,
		"https://management.azure.com/subscriptions/sub-err/resourceGroups/rg-1/providers/Microsoft.Network/virtualNetworks/vnet-1")
	assert.NoError(t, err)

	resp, err := pl.Do(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	assert.Equal(t, float64(1), metricValue(
		"GET /subscriptions/{id}/resourceGroups/{id}/providers/Microsoft.Network/virtualNetworks/{id}",
		"0", "", "sub-err"))
}

func TestMetricsPolicyRegionFromLroPolling(t *testing.T) {
	metrics.CloudProviderCallCount.Reset()

	pl := newMetricsTestPipeline(&fakeTransport{statusCode: http.StatusOK, body: `{"status":"Succeeded"}`})
	req, err := runtime.NewRequest(context.Background(), http.MethodGet,
		"https://management.azure.com/subscriptions/sub-lro/providers/Microsoft.Cache/locations/westeurope/asyncOperations/op-1")
	assert.NoError(t, err)

	_, err = pl.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), metricValue(
		"GET /subscriptions/{id}/providers/Microsoft.Cache/locations/{id}/asyncOperations/{id}",
		"200", "westeurope", "sub-lro"))
}
