package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/alibabacloud-go/tea/dara"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

type accountIdCtxKey struct{}

// AccountIdIntoContext returns a context carrying the AliCloud account id, so the
// metered http client can label calls with a subscription even though AliCloud
// requests never carry the account id. Injected by the state factories before
// they build the SDK client, since the generated AliCloud clients drop the
// per-call context before it reaches the transport.
func AccountIdIntoContext(ctx context.Context, accountId string) context.Context {
	return context.WithValue(ctx, accountIdCtxKey{}, accountId)
}

// AccountIdFromContext returns the AliCloud account id injected by
// AccountIdIntoContext, or an empty string when none was set.
func AccountIdFromContext(ctx context.Context) string {
	if accountId, ok := ctx.Value(accountIdCtxKey{}).(string); ok {
		return accountId
	}
	return ""
}

// AccountIdFromScope returns the AliCloud account id of the given Scope, or an
// empty string when the scope is missing or is not an AliCloud one.
func AccountIdFromScope(scope *cloudcontrolv1beta1.Scope) string {
	if scope == nil || scope.Spec.Scope.Alicloud == nil {
		return ""
	}
	return scope.Spec.Scope.Alicloud.AccountId
}

// AccountIdFromSubscription returns the AliCloud account id of the given
// Subscription, or an empty string when it is missing or is not an AliCloud one.
func AccountIdFromSubscription(subscription *cloudcontrolv1beta1.Subscription) string {
	if subscription == nil ||
		subscription.Status.SubscriptionInfo == nil ||
		subscription.Status.SubscriptionInfo.Alicloud == nil {
		return ""
	}
	return subscription.Status.SubscriptionInfo.Alicloud.AccountId
}

// NewMetricsHTTPClient returns a dara.HttpClient that counts AliCloud API calls
// in the CloudProviderCallCount metric. Assign it to openapi.Config.HttpClient
// when building an AliCloud SDK client.
//
// darabonba-openapi routes every http attempt through this client, so SDK retries
// are counted as separate calls, matching AWS, GCP, Azure and OpenStack. region
// and subscription are fixed per client: the endpoint is region scoped and the
// account id is not part of the request.
func NewMetricsHTTPClient(region, subscription string) dara.HttpClient {
	return &meteredHTTPClient{
		region:       region,
		subscription: subscription,
	}
}

type meteredHTTPClient struct {
	region       string
	subscription string
}

func (c *meteredHTTPClient) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	host := ""
	if request.URL != nil {
		host = request.URL.Host
	}

	response, err := pooledClient(host, transport).Do(request)

	responseCode := 0
	if response != nil {
		responseCode = response.StatusCode
	}

	metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderAliCloud,
		methodLabel(request, host),
		fmt.Sprintf("%d", responseCode),
		c.region,
		c.subscription,
	).Inc()

	return response, err
}

// methodLabel names the api call as <service>/<action>, mirroring the
// <service>/<operation> label AWS reports. AliCloud RPC calls all go to POST /,
// so the action has to be read off the request itself rather than off its path.
func methodLabel(request *http.Request, host string) string {
	action := actionOf(request)
	if action == "" {
		// Not a call the SDK tagged with an action; fall back to the http verb and
		// path so it is still recorded. Both services Cloud Manager calls are RPC
		// style, whose path is always "/", so the label stays bounded.
		path := ""
		if request.URL != nil {
			path = request.URL.Path
		}
		action = fmt.Sprintf("%s %s", request.Method, path)
	}

	service := serviceOfHost(host)
	if service == "" {
		return action
	}

	return service + "/" + action
}

// actionHeader is the header darabonba-openapi tags every call with. The dara
// transport writes SDK headers under their raw lowercase keys rather than the
// canonical ones http.Header.Get assumes, so the header is matched case
// insensitively instead of looked up.
const actionHeader = "x-acs-action"

// actionOf returns the AliCloud api action the request invokes. Calls signed with
// the older v2 algorithm carry the action as a query parameter instead of a
// header.
func actionOf(request *http.Request) string {
	for key, values := range request.Header {
		if len(values) > 0 && strings.EqualFold(key, actionHeader) {
			return values[0]
		}
	}
	if request.URL != nil {
		return request.URL.Query().Get("Action")
	}

	return ""
}

// serviceOfHost returns the AliCloud service an endpoint belongs to. AliCloud
// endpoints are <service>.<region>.aliyuncs.com, so the leading label names the
// service. Hosts not of that shape - a bare address, as tests use - have no
// service to report.
func serviceOfHost(host string) string {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}
	if net.ParseIP(host) != nil {
		return ""
	}

	service, rest, found := strings.Cut(host, ".")
	if !found || rest == "" {
		return ""
	}

	return service
}

// httpClientPool holds one http.Client per endpoint host, so connections survive
// the SDK clients Cloud Manager rebuilds on every reconcile. dara pools its own
// clients per host and hands each one its transport only once; a client installed
// through openapi.Config.HttpClient opts out of that pool, so it has to bring its
// own or every api call would open a fresh connection and leak the idle one - the
// transports dara builds have no idle timeout. Unlike dara's client this one does
// not carry an http.Client timeout, which is equivalent only as long as no
// AliCloud config sets ReadTimeout or ConnectTimeout.
var httpClientPool sync.Map // host -> *http.Client

func pooledClient(host string, transport *http.Transport) *http.Client {
	if client, ok := httpClientPool.Load(host); ok {
		return client.(*http.Client)
	}

	client, _ := httpClientPool.LoadOrStore(host, &http.Client{Transport: transport})

	return client.(*http.Client)
}
