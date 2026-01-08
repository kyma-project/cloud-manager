package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

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
	for i := len(method) - 1; i >= 0; i-- {
		if method[i] == '/' {
			return method[i+1:]
		}
	}
	return method
}

func extractRegionAndProject(req interface{}) (region string, project string) {
	if req == nil {
		return "", ""
	}

	val := reflect.ValueOf(req)

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "", ""
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", ""
	}

	if regionField := val.FieldByName("Region"); regionField.IsValid() && regionField.Kind() == reflect.String {
		region = regionField.String()
	}

	if projectField := val.FieldByName("Project"); projectField.IsValid() && projectField.Kind() == reflect.String {
		project = projectField.String()
	} else if projectIdField := val.FieldByName("ProjectId"); projectIdField.IsValid() && projectIdField.Kind() == reflect.String {
		project = projectIdField.String()
	}

	return region, project
}

// This function is kept for backward compatibility with OLD pattern clients
func IncrementCallCounter(serviceName, operationName, region string, project string, err error) {
	responseCode := 200
	if err != nil {
		// Try to extract response code from googleapi.Error (REST APIs)
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			responseCode = e.Code
		} else if s, ok := status.FromError(err); ok {
			// Extract from gRPC status (modern Cloud Client Libraries)
			responseCode = httpCodeFromGrpcCode(s.Code())
		} else {
			// Generic error
			responseCode = 500
		}
	}
	metrics.CloudProviderCallCount.WithLabelValues(
		metrics.CloudProviderGCP,
		serviceName+"/"+operationName,
		fmt.Sprintf("%d", responseCode),
		region,
		project,
	).Inc()
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
