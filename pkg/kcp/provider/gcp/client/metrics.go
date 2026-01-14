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

// Example: /v1/projects/my-project/locations/europe-west1/instances/my-instance -> ("europe-west1", "my-project")
func extractRegionAndProjectFromURL(path string) (region, project string) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "projects":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				project = parts[i+1]
			}
		case "regions", "locations":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				region = parts[i+1]
			}
		case "zones":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "{") {
				zone := parts[i+1]
				if idx := strings.LastIndexByte(zone, '-'); idx > 0 {
					region = zone[:idx]
				}
			}
		}
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
	region, project := extractRegionAndProjectFromURL(req.URL.Path)
	apiErr := m.convertToAPIError(resp, err)

	IncrementCallCounter(m.serviceName, operation, region, project, apiErr)

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
// Example: GET /v1/services/{service}/connections -> "Connections.List"
// Example: POST /v1/services/{service}/connections:deleteConnection -> "Connections.DeleteConnection"
func extractOperationFromURL(path, method string) string {
	parts := filterPathSegments(strings.Split(strings.TrimPrefix(path, "/"), "/"))
	if len(parts) == 0 {
		return method
	}

	// Handle custom methods (":enable", ":disable", ":deleteConnection", etc.)
	lastPart := parts[len(parts)-1]
	if idx := strings.IndexByte(lastPart, ':'); idx > 0 {
		resource := capitalize(strings.TrimSuffix(lastPart, lastPart[idx:]))
		action := capitalize(lastPart[idx+1:])
		return resource + "." + action
	}

	// Standard REST operations
	resource := capitalize(parts[len(parts)-1])
	action := deriveActionFromMethod(method, hasResourceID(parts))
	return resource + "." + action
}

func filterPathSegments(parts []string) []string {
	var cleaned []string
	skipNext := false

	for i, part := range parts {
		if skipNext {
			skipNext = false
			continue
		}

		if i == 0 && strings.HasPrefix(part, "v") {
			continue
		}

		if part == "projects" || part == "locations" || part == "regions" || part == "zones" {
			skipNext = true
			continue
		}

		if strings.HasPrefix(part, "{") {
			continue
		}

		if part == "services" && i+1 < len(parts) && strings.Contains(parts[i+1], ".googleapis.com") {
			skipNext = true
			continue
		}

		cleaned = append(cleaned, part)
	}

	// Second pass: Google's REST API standard pattern
	var result []string
	for i, part := range cleaned {
		if i%2 == 0 {
			result = append(result, part)
		}
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
