package meta

import (
	"errors"

	"google.golang.org/api/googleapi"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var e *googleapi.Error
	if ok := errors.As(err, &e); ok {
		if e.Code == 404 {
			return true
		}
	}

	return false
}
