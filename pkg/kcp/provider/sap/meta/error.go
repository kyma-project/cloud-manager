package meta

import (
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
)

func NewNotFoundError(msg string) error {
	return &gophercloud.ErrUnexpectedResponseCode{
		BaseError: gophercloud.BaseError{
			Info: msg,
		},
		Actual: http.StatusNotFound,
	}
}

func NewBadRequestError(msg string) error {
	return &gophercloud.ErrUnexpectedResponseCode{
		BaseError: gophercloud.BaseError{
			Info: msg,
		},
		Actual: http.StatusBadRequest,
	}
}
