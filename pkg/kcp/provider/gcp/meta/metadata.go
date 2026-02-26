package meta

import (
	"errors"
	"fmt"
	"net/http"

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
		if googleapierror.Code == http.StatusForbidden {
			return true
		}
	}

	var apiError *apierror.APIError
	if ok := errors.As(err, &apiError); ok {
		if apiError.GRPCStatus().Code() == codes.PermissionDenied {
			return true
		}
		if apiError.HTTPCode() == http.StatusForbidden {
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
		if googleapierror.Code == http.StatusTooManyRequests {
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

func NewNotFoundError(message string, a ...any) error {
	herr := &googleapi.Error{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf(message, a...),
	}
	err, _ := apierror.FromError(herr)
	return err
}

func NewNotAuthorizedError(message string, a ...any) error {
	herr := &googleapi.Error{
		Code:    http.StatusForbidden,
		Message: fmt.Sprintf(message, a...),
	}
	err, _ := apierror.FromError(herr)
	return err
}

func NewTooManyRequestsError(message string, a ...any) error {
	herr := &googleapi.Error{
		Code:    http.StatusTooManyRequests,
		Message: fmt.Sprintf(message, a...),
	}
	err, _ := apierror.FromError(herr)
	return err
}

func NewBadRequestError(message string, a ...any) error {
	herr := &googleapi.Error{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf(message, a...),
	}
	err, _ := apierror.FromError(herr)
	return err
}

func NewInternalServerError(message string, a ...any) error {
	herr := &googleapi.Error{
		Code:    http.StatusInternalServerError,
		Message: fmt.Sprintf(message, a...),
	}
	err, _ := apierror.FromError(herr)
	return err
}
