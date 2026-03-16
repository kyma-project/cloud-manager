package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		region, project := extractFromGrpcContext(ctx)
		IncrementCallCounter(method, region, project, err)
		return err
	}
}

func extractFromGrpcContext(ctx context.Context) (region, project string) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", ""
	}

	if params := md.Get("x-goog-request-params"); len(params) > 0 {
		region, project = parseGoogRequestParams(params[0])
	}

	return region, project
}

func NewMetricsHTTPClient(baseTransport http.RoundTripper) *http.Client {
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	return &http.Client{
		Transport: &metricsRoundTripper{base: baseTransport},
	}
}

type metricsRoundTripper struct {
	base http.RoundTripper
}

func (m *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := m.base.RoundTrip(req)

	var region, project string
	if params := req.Header.Get("x-goog-request-params"); params != "" {
		region, project = parseGoogRequestParams(params)
	}
	if region == "" && project == "" {
		region, project = extractRegionAndProjectFromPath(req.URL.Path)
	}

	sanitizedPath := sanitizePath(req.URL.Path)
	operation := fmt.Sprintf("%s %s", req.Method, sanitizedPath)
	apiErr := m.convertToAPIError(resp, err)

	IncrementCallCounter(operation, region, project, apiErr)

	return resp, err
}

func (m *metricsRoundTripper) convertToAPIError(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode >= 400 {
		return &googleapi.Error{Code: resp.StatusCode}
	}
	return nil
}

func parseGoogRequestParams(params string) (region, project string) {
	for _, pair := range strings.Split(params, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		value, err := url.QueryUnescape(kv[1])
		if err != nil {
			value = kv[1]
		}

		region, project = extractRegionAndProjectFromPath(value)
		if region != "" || project != "" {
			break
		}
	}
	return region, project
}

func extractRegionAndProjectFromPath(path string) (region, project string) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "projects":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				project = sanitizePathValue(parts[i+1])
			}
		case "regions":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				region = sanitizePathValue(parts[i+1])
			}
		case "locations":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				region = convertZoneToRegion(sanitizePathValue(parts[i+1]))
			}
		case "zones":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				region = convertZoneToRegion(sanitizePathValue(parts[i+1]))
			}
		}
	}

	return region, project
}

func sanitizePathValue(value string) string {
	if idx := strings.IndexAny(value, "=&?"); idx != -1 {
		return value[:idx]
	}
	return value
}

func convertZoneToRegion(locationOrZone string) string {
	if len(locationOrZone) > 2 && locationOrZone[len(locationOrZone)-2] == '-' {
		lastChar := locationOrZone[len(locationOrZone)-1]
		if lastChar >= 'a' && lastChar <= 'z' {
			return locationOrZone[:len(locationOrZone)-2]
		}
	}
	return locationOrZone
}

func sanitizePath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var sanitized []string

	for _, part := range parts {
		if part == "" {
			continue
		}

		if strings.Contains(part, ":") {
			idx := strings.Index(part, ":")
			if looksLikeID(part[:idx]) {
				sanitized = append(sanitized, "{id}"+part[idx:])
			} else {
				sanitized = append(sanitized, part)
			}
			continue
		}

		if strings.HasPrefix(part, "v") && len(part) > 1 && part[1] >= '0' && part[1] <= '9' {
			sanitized = append(sanitized, part)
			continue
		}

		if looksLikeID(part) {
			sanitized = append(sanitized, "{id}")
		} else {
			sanitized = append(sanitized, part)
		}
	}

	return "/" + strings.Join(sanitized, "/")
}

func looksLikeID(s string) bool {
	if len(s) == 0 {
		return false
	}

	hasDigit := false
	hasLetter := false
	hasHyphen := false

	for _, c := range s {
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		} else if c == '-' || c == '_' {
			hasHyphen = true
		}
	}

	// Has hyphen - very likely an ID (my-project, res-id, instance-123)
	if hasHyphen && hasLetter {
		return true
	}

	// Short alphanumeric (p1, i1, abc123)
	if hasDigit && hasLetter {
		return true
	}

	return false
}

func IncrementCallCounter(operation, region, project string, err error) {
	code := extractStatusCode(err)
	metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderGCP,
		operation,
		fmt.Sprintf("%d", code),
		region,
		project,
	).Inc()
}

func extractStatusCode(err error) int {
	if err == nil {
		return 200
	}

	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}

	if s, ok := status.FromError(err); ok {
		return httpCodeFromGrpcCode(s.Code())
	}

	return 0
}

func httpCodeFromGrpcCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return 200
	case codes.Canceled:
		return 499
	case codes.Unknown:
		return 500
	case codes.InvalidArgument:
		return 400
	case codes.DeadlineExceeded:
		return 504
	case codes.NotFound:
		return 404
	case codes.AlreadyExists:
		return 409
	case codes.PermissionDenied:
		return 403
	case codes.ResourceExhausted:
		return 429
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 409
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 501
	case codes.Internal:
		return 500
	case codes.Unavailable:
		return 503
	case codes.DataLoss:
		return 500
	case codes.Unauthenticated:
		return 401
	default:
		return 999
	}
}
