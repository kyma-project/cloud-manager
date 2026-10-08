package metrics

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

const pathPlaceholder = "{id}"

type regionCtxKey struct{}

// RegionIntoContext returns a context carrying the Azure region, so the metrics
// policy can label calls with a region even though ARM resource paths do not
// contain one (only LRO/location-scoped urls do). Injected at the provider flow
// entry points from scope.Spec.Region.
func RegionIntoContext(ctx context.Context, region string) context.Context {
	return context.WithValue(ctx, regionCtxKey{}, region)
}

// RegionFromContext returns the Azure region injected by RegionIntoContext, or
// an empty string when none was set.
func RegionFromContext(ctx context.Context) string {
	if region, ok := ctx.Value(regionCtxKey{}).(string); ok {
		return region
	}
	return ""
}

// metricsPolicy counts Azure API calls in the CloudProviderCallCount metric.
// It is injected as a PerRetryPolicy, so it runs below the SDK retry policy and
// observes every HTTP attempt rather than only the outcome of the SDK call.
// azcore retries 408, 429 and 5xx three times by default, so counting above the
// retry policy would report a throttled-then-successful call as a single 200 and
// hide the 429s entirely. Counting per attempt also matches how AWS, GCP and
// OpenStack are instrumented, keeping the metric comparable across providers.
type metricsPolicy struct{}

var metricsPolicyInstance = &metricsPolicy{}

// NewMetricsPolicy returns the policy counting Azure API calls in the
// CloudProviderCallCount metric. The policy is stateless, so all clients
// share a single instance.
func NewMetricsPolicy() policy.Policy {
	return metricsPolicyInstance
}

func (p *metricsPolicy) Do(req *policy.Request) (*http.Response, error) {
	resp, err := req.Next()

	responseCode := 0
	if resp != nil {
		responseCode = resp.StatusCode
	}

	sanitizedPath, urlRegion, subscription := parseAzureRequestPath(req.Raw().URL.Path)

	// Prefer the region injected into the context by the reconciler, since ARM
	// resource paths carry a region only on LRO/location-scoped urls. Fall back
	// to the url-parsed region so those calls still get labeled when no region
	// was injected.
	region := RegionFromContext(req.Raw().Context())
	if region == "" {
		region = urlRegion
	}

	metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderAzure,
		fmt.Sprintf("%s %s", req.Raw().Method, sanitizedPath),
		fmt.Sprintf("%d", responseCode),
		region,
		subscription,
	).Inc()

	return resp, err
}

// parseAzureRequestPath masks the variable segments of an ARM request path
// with {id} so the method label cardinality stays bounded, and extracts the
// subscription id and, when the path contains a locations segment (as LRO
// polling urls do), the region. ARM paths follow
// /subscriptions/{id}/resourceGroups/{name}/providers/{Namespace} followed by
// alternating {resourceType}/{resourceName} segments.
func parseAzureRequestPath(path string) (sanitizedPath, region, subscription string) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "/", "", ""
	}

	const (
		rootKeyword = iota
		rootValue
		providerNamespace
		resourceType
		resourceName
	)

	captureSubscription := false
	captureRegion := false

	segments := strings.Split(trimmed, "/")
	sanitized := make([]string, len(segments))

	mode := rootKeyword
	for i, segment := range segments {
		switch mode {
		case rootKeyword:
			sanitized[i] = keepSegment(segment)
			captureSubscription = false
			captureRegion = false
			switch {
			case strings.EqualFold(segment, "subscriptions"):
				sanitized[i] = "subscriptions"
				captureSubscription = true
				mode = rootValue
			case strings.EqualFold(segment, "resourceGroups"):
				sanitized[i] = "resourceGroups"
				mode = rootValue
			case strings.EqualFold(segment, "locations"):
				sanitized[i] = "locations"
				captureRegion = true
				mode = rootValue
			case strings.EqualFold(segment, "providers"):
				sanitized[i] = "providers"
				mode = providerNamespace
			}
		case rootValue:
			sanitized[i] = pathPlaceholder
			if captureSubscription {
				subscription = segment
			}
			if captureRegion {
				region = segment
			}
			mode = rootKeyword
		case providerNamespace:
			sanitized[i] = segment
			mode = resourceType
		case resourceType:
			sanitized[i] = keepSegment(segment)
			captureRegion = false
			switch {
			case strings.EqualFold(segment, "providers"):
				sanitized[i] = "providers"
				mode = providerNamespace
			case strings.EqualFold(segment, "locations"):
				sanitized[i] = "locations"
				captureRegion = true
				mode = resourceName
			default:
				mode = resourceName
			}
		case resourceName:
			sanitized[i] = pathPlaceholder
			if captureRegion {
				region = segment
			}
			mode = resourceType
		}
	}

	return "/" + strings.Join(sanitized, "/"), region, subscription
}

var uuidRegexp = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// keepSegment protects against uuids landing on a keyword/type position, as
// in the operationResults urls some LRO pollers follow, where the alternating
// type/name grammar does not hold.
func keepSegment(segment string) string {
	if uuidRegexp.MatchString(segment) {
		return pathPlaceholder
	}
	return segment
}
