package meta

import (
	"errors"
	"testing"

	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundResponseError(t *testing.T) {
	err := NewHttpNotFoundError(errors.New("not found error"))
	assert.True(t, IsNotFound(err))
}

func TestIsNotFound(t *testing.T) {
	codes := []string{
		"FileSystemNotFound",
		"AccessPointNotFound",
		"MountTargetNotFound",
		"PolicyNotFound",
		"CacheSubnetGroupNotFoundFault",
		"CacheClusterNotFound",
		"ResourceNotFoundException",
		"InvalidVpcID.NotFound",
		"InvalidVpcPeeringConnectionID.NotFound",
		"InvalidRoute.NotFound",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			assert.True(t, IsNotFound(&smithy.GenericAPIError{Code: code}))
		})
	}
}

func TestIsInvalidParameter(t *testing.T) {
	t.Run("returns true for InvalidParameterValueException", func(t *testing.T) {
		err := &backuptypes.InvalidParameterValueException{
			Message: strPtr("Idempotency token already used"),
		}
		assert.True(t, IsInvalidParameter(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsInvalidParameter(nil))
	})

	t.Run("returns false for other error types", func(t *testing.T) {
		err := errors.New("some other error")
		assert.False(t, IsInvalidParameter(err))
	})

	t.Run("returns false for wrapped non-matching error", func(t *testing.T) {
		err := &backuptypes.ResourceNotFoundException{
			Message: strPtr("not found"),
		}
		assert.False(t, IsInvalidParameter(err))
	})
}

func strPtr(s string) *string {
	return &s
}
