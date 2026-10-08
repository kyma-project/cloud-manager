package client

import (
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
)

func NewNotFoundError(msg string) error {
	result := &gophercloud.ErrUnexpectedResponseCode{
		Info:     msg,
		Expected: []int{http.StatusOK},
		Actual:   http.StatusNotFound,
	}
	return result
}
