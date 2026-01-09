package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/metrics"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor creates a gRPC unary interceptor that automatically meters API calls
// to the cloud_manager_cloud_provider_api_call_total metric.
func UnaryClientInterceptor(serviceName, resource string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		operation := extractOperationName(method)
		region, project := extractRegionAndProject(req)

		IncrementCallCounter(serviceName, resource+"."+operation, region, project, err)

		return err
	}
}

// Example: "/google.cloud.compute.v1.Addresses/Insert" -> "Insert"
func extractOperationName(method string) string {
	if idx := strings.LastIndexByte(method, '/'); idx >= 0 {
		return method[idx+1:]
	}
	return method
}

func extractRegionAndProject(req interface{}) (region, project string) {
	if req == nil {
		return "", ""
	}

	val := reflect.ValueOf(req)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "", ""
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", ""
	}

	if field := val.FieldByName("Region"); field.IsValid() && field.Kind() == reflect.String {
		region = field.String()
	}

	if field := val.FieldByName("Project"); field.IsValid() && field.Kind() == reflect.String {
		project = field.String()
	} else if field := val.FieldByName("ProjectId"); field.IsValid() && field.Kind() == reflect.String {
		project = field.String()
	}

	return region, project
}

// metricsRoundTripper wraps an HTTP transport to automatically record metrics for REST API calls
type metricsRoundTripper struct {
	base        http.RoundTripper
	serviceName string
}

// RoundTrip implements http.RoundTripper interface
func (m *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := m.base.RoundTrip(req)

	operation := extractOperationFromURL(req.URL.Path, req.Method)
	apiErr := m.convertToAPIError(resp, err)

	IncrementCallCounter(m.serviceName, operation, "", "", apiErr)

	return resp, err
}

// convertToAPIError normalizes response errors for metric recording
func (m *metricsRoundTripper) convertToAPIError(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp != nil && resp.StatusCode >= 400 {
		return &googleapi.Error{Code: resp.StatusCode}
	}
	return nil
}

// extractOperationFromURL extracts the operation name from REST API URL path
// Example: POST /v1/projects/{project}/services/{service}:enable -> "Services.Enable"
func extractOperationFromURL(path, method string) string {
	parts := filterPathSegments(strings.Split(strings.TrimPrefix(path, "/"), "/"))
	if len(parts) == 0 {
		return method
	}

	// Handle custom methods (":enable", ":disable", etc.)
	if idx := strings.IndexByte(parts[len(parts)-1], ':'); idx > 0 {
		if len(parts) >= 2 {
			resource := capitalize(parts[len(parts)-2])
			action := capitalize(parts[len(parts)-1][idx+1:])
			return resource + "." + action
		}
	}

	// Standard REST operations
	resource := capitalize(parts[0])
	action := deriveActionFromMethod(method, hasResourceID(parts))
	return resource + "." + action
}

// filterPathSegments removes version, project, and placeholder segments
func filterPathSegments(parts []string) []string {
	var result []string
	skipNext := false

	for i, part := range parts {
		if skipNext {
			skipNext = false
			continue
		}

		// Skip version (v1, v1beta1) and "projects" at start
		if i == 0 && (strings.HasPrefix(part, "v") || part == "projects") {
			continue
		}

		// Skip "projects" and "{...}" placeholders along with next segment
		if part == "projects" || part == "locations" || strings.HasPrefix(part, "{") {
			skipNext = true
			continue
		}

		result = append(result, part)
	}

	return result
}

// hasResourceID checks if the path targets a specific resource
func hasResourceID(parts []string) bool {
	return len(parts) > 1 && !strings.HasPrefix(parts[len(parts)-1], "{")
}

// deriveActionFromMethod maps HTTP methods to action names
func deriveActionFromMethod(method string, hasID bool) string {
	if hasID {
		switch method {
		case "GET":
			return "Get"
		case "DELETE":
			return "Delete"
		case "PUT", "PATCH":
			return "Update"
		}
	} else {
		switch method {
		case "GET":
			return "List"
		case "POST":
			return "Insert"
		}
	}
	return method
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// NewMetricsHTTPClient creates an HTTP client with automatic metrics recording for REST APIs
func NewMetricsHTTPClient(serviceName string, baseTransport http.RoundTripper) *http.Client {
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	return &http.Client{
		Transport: &metricsRoundTripper{
			base:        baseTransport,
			serviceName: serviceName,
		},
	}
}

// IncrementCallCounter records API call metrics
func IncrementCallCounter(serviceName, operationName, region, project string, err error) {
	code := extractHTTPStatusCode(err)
	metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderGCP,
		serviceName+"/"+operationName,
		fmt.Sprintf("%d", code),
		region,
		project,
	).Inc()
}

// extractHTTPStatusCode converts errors to HTTP status codes
func extractHTTPStatusCode(err error) int {
	if err == nil {
		return 200
	}

	// Try googleapi.Error (REST APIs)
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}

	// Try gRPC status (gRPC APIs)
	if s, ok := status.FromError(err); ok {
		return httpCodeFromGrpcCode(s.Code())
	}

	// Generic error
	return 500
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
		return 500
	}
}
