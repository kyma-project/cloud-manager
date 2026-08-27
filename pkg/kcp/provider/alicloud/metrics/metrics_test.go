package metrics

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	vpc "github.com/alibabacloud-go/vpc-20160428/v6/client"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestServiceOfHost(t *testing.T) {
	testCases := []struct {
		name            string
		host            string
		expectedService string
	}{
		{name: "regional vpc endpoint", host: "vpc.eu-central-1.aliyuncs.com", expectedService: "vpc"},
		{name: "regional nas endpoint", host: "nas.eu-central-1.aliyuncs.com", expectedService: "nas"},
		{name: "central endpoint", host: "vpc.aliyuncs.com", expectedService: "vpc"},
		{name: "endpoint with port", host: "vpc.eu-central-1.aliyuncs.com:443", expectedService: "vpc"},
		{name: "ipv4 address", host: "127.0.0.1:8080", expectedService: ""},
		{name: "ipv6 address", host: "[::1]:8080", expectedService: ""},
		{name: "single label host", host: "localhost", expectedService: ""},
		{name: "trailing dot only", host: "vpc.", expectedService: ""},
		{name: "empty host", host: "", expectedService: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedService, serviceOfHost(tc.host))
		})
	}
}

func TestMethodLabel(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		httpMethod    string
		header        http.Header
		expectedLabel string
	}{
		{
			name:       "action from the raw lowercase header the sdk writes",
			url:        "http://vpc.eu-central-1.aliyuncs.com/",
			httpMethod: http.MethodPost,
			// The dara transport assigns SDK headers under their raw key, so this is
			// how the header actually reaches the metered client.
			header:        http.Header{"x-acs-action": []string{"DescribeVpcs"}},
			expectedLabel: "vpc/DescribeVpcs",
		},
		{
			name:          "action from the canonical header",
			url:           "http://nas.eu-central-1.aliyuncs.com/",
			httpMethod:    http.MethodPost,
			header:        http.Header{"X-Acs-Action": []string{"CreateFileSystem"}},
			expectedLabel: "nas/CreateFileSystem",
		},
		{
			name:          "action from the query of a v2 signed call",
			url:           "http://vpc.eu-central-1.aliyuncs.com/?Action=DescribeVSwitches&Version=2016-04-28",
			httpMethod:    http.MethodPost,
			expectedLabel: "vpc/DescribeVSwitches",
		},
		{
			name:          "untagged call falls back to the http verb and path",
			url:           "http://vpc.eu-central-1.aliyuncs.com/",
			httpMethod:    http.MethodGet,
			expectedLabel: "vpc/GET /",
		},
		{
			name:          "host without a service reports the action alone",
			url:           "http://127.0.0.1:8080/",
			httpMethod:    http.MethodPost,
			header:        http.Header{"x-acs-action": []string{"DescribeVpcs"}},
			expectedLabel: "DescribeVpcs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request, err := http.NewRequest(tc.httpMethod, tc.url, nil)
			assert.NoError(t, err)
			for key, values := range tc.header {
				request.Header[key] = values
			}

			assert.Equal(t, tc.expectedLabel, methodLabel(request, request.URL.Host))
		})
	}
}

func TestAccountIdContextRoundTrip(t *testing.T) {
	ctx := AccountIdIntoContext(context.Background(), "1234567890")
	assert.Equal(t, "1234567890", AccountIdFromContext(ctx))
}

func TestAccountIdFromContextEmptyWhenAbsent(t *testing.T) {
	assert.Equal(t, "", AccountIdFromContext(context.Background()))
}

func TestAccountIdFromScope(t *testing.T) {
	scope := &cloudcontrolv1beta1.Scope{
		Spec: cloudcontrolv1beta1.ScopeSpec{
			Scope: cloudcontrolv1beta1.ScopeInfo{
				Alicloud: &cloudcontrolv1beta1.AlicloudScope{AccountId: "1234567890"},
			},
		},
	}
	assert.Equal(t, "1234567890", AccountIdFromScope(scope))

	assert.Equal(t, "", AccountIdFromScope(nil))
	assert.Equal(t, "", AccountIdFromScope(&cloudcontrolv1beta1.Scope{}))
}

func TestAccountIdFromSubscription(t *testing.T) {
	subscription := &cloudcontrolv1beta1.Subscription{
		Status: cloudcontrolv1beta1.SubscriptionStatus{
			SubscriptionInfo: &cloudcontrolv1beta1.SubscriptionInfo{
				Alicloud: &cloudcontrolv1beta1.SubscriptionInfoAlicloud{AccountId: "1234567890"},
			},
		},
	}
	assert.Equal(t, "1234567890", AccountIdFromSubscription(subscription))

	assert.Equal(t, "", AccountIdFromSubscription(nil))
	assert.Equal(t, "", AccountIdFromSubscription(&cloudcontrolv1beta1.Subscription{}))
	assert.Equal(t, "", AccountIdFromSubscription(&cloudcontrolv1beta1.Subscription{
		Status: cloudcontrolv1beta1.SubscriptionStatus{
			SubscriptionInfo: &cloudcontrolv1beta1.SubscriptionInfo{},
		},
	}))
}

const (
	testEndpoint  = "vpc.eu-central-1.aliyuncs.com"
	testRegion    = "eu-central-1"
	testAccountId = "1234567890"
)

func metricValue(method, responseCode, region, subscription string) float64 {
	return testutil.ToFloat64(metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderAliCloud,
		method,
		responseCode,
		region,
		subscription,
	))
}

// resetMetrics clears both the counter and the http client pool, so a test is not
// served the client another test pooled for the same endpoint host.
func resetMetrics() {
	metrics.CloudProviderCallCount.Reset()
	httpClientPool.Range(func(key, _ any) bool {
		httpClientPool.Delete(key)
		return true
	})
}

// newRoutingTransport returns a transport dialing the given test server whatever
// host is requested, so requests can carry a realistic AliCloud endpoint host.
func newRoutingTransport(server *httptest.Server) *http.Transport {
	address := server.Listener.Addr().String()
	return &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, address)
		},
	}
}

func newTestRequest(t *testing.T, action string) *http.Request {
	request, err := http.NewRequest(http.MethodPost, "http://"+testEndpoint+"/", nil)
	assert.NoError(t, err)
	request.Header["x-acs-action"] = []string{action}
	return request
}

func TestMeteredHTTPClientSuccessfulCall(t *testing.T) {
	resetMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewMetricsHTTPClient(testRegion, testAccountId)
	response, err := client.Call(newTestRequest(t, "DescribeVpcs"), newRoutingTransport(server))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.NoError(t, response.Body.Close())

	assert.Equal(t, float64(1), metricValue("vpc/DescribeVpcs", "200", testRegion, testAccountId))
}

func TestMeteredHTTPClientFailedCall(t *testing.T) {
	resetMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewMetricsHTTPClient(testRegion, testAccountId)
	response, err := client.Call(newTestRequest(t, "DescribeVSwitches"), newRoutingTransport(server))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	assert.NoError(t, response.Body.Close())

	assert.Equal(t, float64(1), metricValue("vpc/DescribeVSwitches", "404", testRegion, testAccountId))
}

func TestMeteredHTTPClientThrottledCall(t *testing.T) {
	resetMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewMetricsHTTPClient(testRegion, testAccountId)
	response, err := client.Call(newTestRequest(t, "CreateVSwitch"), newRoutingTransport(server))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	assert.NoError(t, response.Body.Close())

	assert.Equal(t, float64(1), metricValue("vpc/CreateVSwitch", "429", testRegion, testAccountId))
}

// A call that never reached AliCloud has no status code to report, so it is
// recorded as 0 - the same as AWS, Azure and OpenStack do.
func TestMeteredHTTPClientTransportError(t *testing.T) {
	resetMetrics()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, errors.New("connection refused")
		},
	}

	client := NewMetricsHTTPClient(testRegion, testAccountId)
	response, err := client.Call(newTestRequest(t, "DescribeVpcs"), transport)
	assert.Error(t, err)
	assert.Nil(t, response)

	assert.Equal(t, float64(1), metricValue("vpc/DescribeVpcs", "0", testRegion, testAccountId))
}

// Each attempt is counted separately, so the SDK's own retries of a throttled
// call show up in the metric instead of collapsing into its final outcome.
func TestMeteredHTTPClientCountsEveryAttempt(t *testing.T) {
	resetMetrics()

	statusCodes := []int{http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusOK}
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCodes[attempts])
		attempts++
	}))
	defer server.Close()

	client := NewMetricsHTTPClient(testRegion, testAccountId)
	transport := newRoutingTransport(server)
	for range statusCodes {
		response, err := client.Call(newTestRequest(t, "CreateVpc"), transport)
		assert.NoError(t, err)
		assert.NoError(t, response.Body.Close())
	}

	assert.Equal(t, float64(2), metricValue("vpc/CreateVpc", "429", testRegion, testAccountId))
	assert.Equal(t, float64(1), metricValue("vpc/CreateVpc", "200", testRegion, testAccountId))
}

// TestMeteredHTTPClientRecordsSdkCall pins the assumptions the instrumentation
// rests on: that darabonba-openapi routes its calls through the http client set
// on openapi.Config, and that the action it tags them with is readable from the
// request that arrives at the transport.
func TestMeteredHTTPClientRecordsSdkCall(t *testing.T) {
	resetMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"RequestId":"req-1","TotalCount":0,"Vpcs":{"Vpc":[]}}`))
	}))
	defer server.Close()

	// The SDK addresses the real AliCloud endpoint, so the client pooled for that
	// host is seeded up front with a transport dialing the test server instead.
	httpClientPool.Store(testEndpoint, &http.Client{Transport: newRoutingTransport(server)})

	config := &openapi.Config{
		AccessKeyId:     new("test-access-key-id"),
		AccessKeySecret: new("test-access-key-secret"),
		RegionId:        new(testRegion),
		Protocol:        new("HTTP"),
	}
	config.Endpoint = new(testEndpoint)
	config.HttpClient = NewMetricsHTTPClient(testRegion, testAccountId)

	vpcClient, err := vpc.NewClient(config)
	assert.NoError(t, err)

	_, err = vpcClient.DescribeVpcs(&vpc.DescribeVpcsRequest{RegionId: new(testRegion)})
	assert.NoError(t, err)

	assert.Equal(t, float64(1), metricValue("vpc/DescribeVpcs", "200", testRegion, testAccountId))
}
