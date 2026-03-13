package meta

import (
	"errors"
	"strings"

	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var googleApiError *googleapi.Error
	if ok := errors.As(err, &googleApiError); ok {
		if googleApiError.Code == 404 {
			return true
		}
	}

	var apiError *apierror.APIError
	if ok := errors.As(err, &apiError); ok {
		if apiError.GRPCStatus().Code() == codes.NotFound {
			return true
		}
	}

	return false
}

func IsNotAuthorized(err error) bool {
	if err == nil {
		return false
	}

	var googleapierror *googleapi.Error
	if ok := errors.As(err, &googleapierror); ok {
		if googleapierror.Code == 403 {
			return true
		}
	}

	var apiError *apierror.APIError
	if ok := errors.As(err, &apiError); ok {
		if apiError.GRPCStatus().Code() == codes.PermissionDenied {
			return true
		}
	}

	return false
}

func IsTooManyRequests(err error) bool {
	if err == nil {
		return false
	}

	var googleapierror *googleapi.Error
	if ok := errors.As(err, &googleapierror); ok {
		if googleapierror.Code == 429 {
			return true
		}
	}

	var apiError *apierror.APIError
	if ok := errors.As(err, &apiError); ok {
		// ResourceExhausted is the gRPC code for TooManyRequests
		if apiError.GRPCStatus().Code() == codes.ResourceExhausted {
			return true
		}
	}

	return false
}

func IsOperationInProgressError(err error) bool {
	if err == nil {
		return false
	}

	var googleapierror *googleapi.Error
	if ok := errors.As(err, &googleapierror); ok {
		if googleapierror.Code == 400 {
			for _, e := range googleapierror.Errors {
				if strings.Contains(e.Message, "operation in progress") {
					return true
				}
			}
		}
	}
	return false
}
