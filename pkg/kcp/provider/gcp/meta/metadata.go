package meta

import (
	"errors"

	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var googleapierror *googleapi.Error
	if ok := errors.As(err, &googleapierror); ok {
		if googleapierror.Code == 404 {
			return true
		}
	}

	var apierror *apierror.APIError
	if ok := errors.As(err, &apierror); ok {
		if apierror.GRPCStatus().Code() == codes.NotFound {
			return true
		}
	}

	return false
}
